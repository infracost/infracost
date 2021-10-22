package terraform

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var cacheFileVersion = "0.1"
var infracostDir = ".infracost"
var cacheFileName = path.Join(infracostDir, ".infracost-cache")
var cacheMaxAgeSecs int64 = 60 * 10 // 10 minutes

type terraformConfigFileState struct {
	Filepath string `json:"filepath"`
	Modified string `json:"modified"`
}

type configState struct {
	Version             string                     `json:"version"`
	TerraformPlanFlags  string                     `json:"terraform_plan_flags"`
	TerraformUseState   bool                       `json:"terraform_use_state"`
	TerraformWorkspace  string                     `json:"terraform_workspace"`
	TerraformBinary     string                     `json:"terraform_binary"`
	TerraformCloudToken string                     `json:"terraform_cloud_token"`
	TerraformCloudHost  string                     `json:"terraform_cloud_host"`
	ConfigEnv           string                     `json:"config_env"`
	TFEnv               string                     `json:"tf_env"`
	TFConfigFileStates  []terraformConfigFileState `json:"tf_config_file_states"`
}

func (state *configState) equivalent(otherState *configState) bool {
	if state.Version != otherState.Version {
		log.Debugf("Plan cache config state not equivalent: version changed")
		return false
	}

	if state.TerraformPlanFlags != otherState.TerraformPlanFlags {
		log.Debugf("Plan cache config state not equivalent: terraform_plan_flags changed")
		return false
	}

	if state.TerraformUseState != otherState.TerraformUseState {
		log.Debugf("Plan cache config state not equivalent: terraform_use_state changed")
		return false
	}

	if state.TerraformWorkspace != otherState.TerraformWorkspace {
		log.Debugf("Plan cache config state not equivalent: terraform_workspace changed")
		return false
	}

	if state.TerraformBinary != otherState.TerraformBinary {
		log.Debugf("Plan cache config state not equivalent: terraform_binary changed")
		return false
	}

	if state.TerraformCloudToken != otherState.TerraformCloudToken {
		log.Debugf("Plan cache config state not equivalent: terraform_cloud_token changed")
		return false
	}

	if state.TerraformCloudHost != otherState.TerraformCloudHost {
		log.Debugf("Plan cache config state not equivalent: terraform_cloud_host changed")
		return false
	}

	if state.ConfigEnv != otherState.ConfigEnv {
		log.Debugf("Plan cache config state not equivalent: config_env changed")
		return false
	}

	if state.TFEnv != otherState.TFEnv {
		log.Debugf("Plan cache config state not equivalent: tf_env changed")
		return false
	}

	if len(state.TFConfigFileStates) != len(otherState.TFConfigFileStates) {
		log.Debugf("Plan cache config state not equivalent: TFConfigFileStates has changed size")
		return false
	}

	for i := range state.TFConfigFileStates {
		if state.TFConfigFileStates[i] != otherState.TFConfigFileStates[i] {
			log.Debugf("Plan cache config state not equivalent: %v", state.TFConfigFileStates[i])
			return false
		}
	}

	return true
}

type cacheFile struct {
	ConfigState configState `json:"config_state"`
	Plan        []byte      `json:"plan"`
}

func ReadPlanCache(p *DirProvider) []byte {
	cache := path.Join(p.Path, cacheFileName)

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

	state := calcConfigState(p)
	if !cf.ConfigState.equivalent(&state) {
		log.Debugf("Skipping plan cache: Config state has changed")
		return nil
	}

	log.Debugf("Read plan JSON from %v", cacheFileName)
	return cf.Plan
}

func WritePlanCache(p *DirProvider, planJSON []byte) {
	cacheJSON, err := json.Marshal(cacheFile{ConfigState: calcConfigState(p), Plan: planJSON})
	if err != nil {
		log.Debugf("Failed to marshal plan cache: %v", err)
		return
	}

	// create the .infracost dir if it doesn't already exist
	if _, err := os.Stat(path.Join(p.Path, infracostDir)); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(path.Join(p.Path, infracostDir), 0700)
			if err != nil {
				log.Debugf("Couldn't create %v directory: %v", infracostDir, err)
				return
			}
		}
	}

	err = os.WriteFile(path.Join(p.Path, cacheFileName), cacheJSON, 0600)
	if err != nil {
		log.Debugf("Failed to write plan cache: %v", err)
		return
	}
	log.Debugf("Wrote plan JSON to %v", cacheFileName)
}

func calcConfigState(p *DirProvider) configState {
	return configState{
		Version:             cacheFileVersion,
		TerraformPlanFlags:  p.PlanFlags,
		TerraformUseState:   p.UseState,
		TerraformWorkspace:  p.Workspace,
		TerraformBinary:     p.TerraformBinary,
		TerraformCloudToken: p.TerraformCloudToken,
		TerraformCloudHost:  p.TerraformCloudHost,
		ConfigEnv:           envToString(p.Env),
		TFEnv:               tfEnvToString(),
		TFConfigFileStates:  calcTerraformConfigFileStates(p.Path),
	}
}

func envToString(env map[string]string) string {
	envPairs := make([]string, 0, len(env))
	for k, v := range env {
		envPairs = append(envPairs, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(envPairs)
	return strings.Join(envPairs, ",")
}

func tfEnvToString() string {
	tfEnvs := []string{}
	for _, s := range os.Environ() {
		if strings.HasPrefix(s, "TF_VAR_") ||
			strings.HasPrefix(s, "TF_CLI_ARGS") ||
			strings.HasPrefix(s, "TF_WORKSPACE") ||
			strings.HasPrefix(s, "TF_CLI_CONFIG_FILE") {
			tfEnvs = append(tfEnvs, s)
		}
	}
	sort.Strings(tfEnvs)
	return strings.Join(tfEnvs, ",")
}

// Finds all files used by a terraform project directory and returns them with their last modified time.
func calcTerraformConfigFileStates(dir string) []terraformConfigFileState {
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
