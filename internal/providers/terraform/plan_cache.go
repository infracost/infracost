package terraform

import (
	"encoding/json"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
)

var cacheFileName = ".infracache"
var cacheMaxAgeSecs int64 = 60 * 10 // 10 minutes

type terraformConfigFileState struct {
	Filepath string `json:"filepath"`
	Modified string `json:"modified"`
}

type cacheFile struct {
	ConfigState []terraformConfigFileState `json:"config_state"`
	Plan        []byte                     `json:"plan"`
}

func ReadPlanCache(dir string) []byte {
	cache := path.Join(dir, cacheFileName)

	info, err := os.Stat(cache)
	if err != nil {
		log.Debugf("Skipping plan cache: Cache file does not exist")
		return nil
	}

	if time.Now().Unix()-info.ModTime().Unix() > cacheMaxAgeSecs {
		log.Debugf("Skipping plan cache: Cache file is too old")
		return nil
	}

	data, err := os.ReadFile(cache)
	if err != nil {
		log.Debugf("Skipping plan cache: Error reading cache file: %v", err)
		return nil
	}

	var cf cacheFile
	err = json.Unmarshal(data, &cf)
	if err != nil {
		log.Debugf("Skipping plan cache: Error unmarshalling cache file: %v", err)
		return nil
	}

	state := calcConfigFileStates(dir)

	if len(state) != len(cf.ConfigState) {
		log.Debugf("Skipping plan cache: ConfigFileState has changed size")
		return nil
	}

	for i := range state {
		if state[i] != cf.ConfigState[i] {
			log.Debugf("Skipping plan cache: ConfigFileState has changed: %v", state[i])
			return nil
		}
	}

	log.Debugf("Read plan JSON from %v", cacheFileName)
	return cf.Plan
}

func WritePlanCache(dir string, planJSON []byte) {
	state := calcConfigFileStates(dir)
	cacheJSON, err := json.Marshal(cacheFile{ConfigState: state, Plan: planJSON})
	if err != nil {
		log.Debugf("Failed to marshal plan cache: %v", err)
		return
	}

	err = os.WriteFile(path.Join(dir, cacheFileName), cacheJSON, 0600)
	if err != nil {
		log.Debugf("Failed to write plan cache: %v", err)
		return
	}
	log.Debugf("Wrote plan JSON to %v", cacheFileName)
}

// Finds all files used by a terraform project directory and returns them with their last modified time.
func calcConfigFileStates(dir string) []terraformConfigFileState {
	filepaths := findNestedSourceFiles(dir)
	configFileStates := make([]terraformConfigFileState, 0, len(filepaths))

	for _, filepath := range filepaths {
		var mod string

		stat, err := os.Stat(filepath)
		if err != nil {
			mod = err.Error()
		} else {
			mod = stat.ModTime().String()
		}

		configFileStates = append(configFileStates, terraformConfigFileState{Filepath: filepath, Modified: mod})
	}

	return configFileStates
}

func findNestedSourceFiles(dir string) []string {
	dirToFiles := make(map[string][]string)

	findSources(dir, dirToFiles)

	// get the unique list of filepaths
	m := make(map[string]bool)
	for _, filepaths := range dirToFiles {
		for _, filepath := range filepaths {
			m[filepath] = true
		}
	}

	allFilepaths := make([]string, 0, len(m))
	for filepath := range m {
		allFilepaths = append(allFilepaths, filepath)
	}

	sort.Strings(allFilepaths)

	return allFilepaths
}

// recursive part of findNestedSourceFiles that populates the dirToFiles map with the filenames used by each module
func findSources(dir string, dirToFiles map[string][]string) {
	if _, ok := dirToFiles[dir]; ok {
		// we have already processed this directory
		return
	}

	module, _ := tfconfig.LoadModule(dir)

	// get all the source files used in this directory
	dirToFiles[dir] = findSourceFiles(module)

	// recursively process any directories used in module/provider blocks.
	for _, m := range module.ModuleCalls {
		findSources(filepath.Join(dir, m.Source), dirToFiles)
	}

	for _, r := range module.RequiredProviders {
		findSources(filepath.Join(dir, r.Source), dirToFiles)
	}
}

func findSourceFiles(module *tfconfig.Module) []string {
	m := make(map[string]bool)

	for _, el := range module.Variables {
		m[el.Pos.Filename] = true
	}

	for _, el := range module.Outputs {
		m[el.Pos.Filename] = true
	}

	for _, el := range module.ManagedResources {
		m[el.Pos.Filename] = true
	}

	for _, el := range module.DataResources {
		m[el.Pos.Filename] = true
	}

	for _, el := range module.ModuleCalls {
		m[el.Pos.Filename] = true
	}

	filepaths := make([]string, 0, len(m))
	for filepath := range m {
		filepaths = append(filepaths, filepath)
	}
	return filepaths
}
