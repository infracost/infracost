package terraform

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
)

var cacheFileVersion = "0.1"
var cacheFileName = ".infracost-cache"
var cacheMaxAgeSecs int64 = 60 * 30 // 30 minutes

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
	TFLockFileDate      string                     `json:"tf_lock_file_date"`
	TFDataDate          string                     `json:"tf_data_date"`
	TFConfigFileStates  []terraformConfigFileState `json:"tf_config_file_states"`
}

func (state *configState) equivalent(otherState *configState) (bool, error) {
	if state.Version != otherState.Version {
		logging.Logger.Debug().Msgf("Plan cache config state not equivalent: version changed")
		return false, fmt.Errorf("version changed")
	}

	if state.TerraformPlanFlags != otherState.TerraformPlanFlags {
		logging.Logger.Debug().Msgf("Plan cache config state not equivalent: terraform_plan_flags changed")
		return false, fmt.Errorf("terraform_plan_flags changed")
	}

	if state.TerraformUseState != otherState.TerraformUseState {
		logging.Logger.Debug().Msgf("Plan cache config state not equivalent: terraform_use_state changed")
		return false, fmt.Errorf("terraform_use_state changed")
	}

	if state.TerraformWorkspace != otherState.TerraformWorkspace {
		logging.Logger.Debug().Msgf("Plan cache config state not equivalent: terraform_workspace changed")
		return false, fmt.Errorf("terraform_workspace changed")
	}

	if state.TerraformBinary != otherState.TerraformBinary {
		logging.Logger.Debug().Msgf("Plan cache config state not equivalent: terraform_binary changed")
		return false, fmt.Errorf("terraform_binary changed")
	}

	if state.TerraformCloudToken != otherState.TerraformCloudToken {
		logging.Logger.Debug().Msgf("Plan cache config state not equivalent: terraform_cloud_token changed")
		return false, fmt.Errorf("terraform_cloud_token changed")
	}

	if state.TerraformCloudHost != otherState.TerraformCloudHost {
		logging.Logger.Debug().Msgf("Plan cache config state not equivalent: terraform_cloud_host changed")
		return false, fmt.Errorf("terraform_cloud_host changed")
	}

	if state.ConfigEnv != otherState.ConfigEnv {
		logging.Logger.Debug().Msgf("Plan cache config state not equivalent: config_env changed")
		return false, fmt.Errorf("config_env changed")
	}

	if state.TFEnv != otherState.TFEnv {
		logging.Logger.Debug().Msgf("Plan cache config state not equivalent: tf_env changed")
		return false, fmt.Errorf("tf_env changed")
	}

	if state.TFLockFileDate != otherState.TFLockFileDate {
		logging.Logger.Debug().Msgf("Plan cache config state not equivalent: tf_lock_file_date changed")
		return false, fmt.Errorf("tf_lock_file_date changed")
	}

	if state.TFDataDate != otherState.TFDataDate {
		logging.Logger.Debug().Msgf("Plan cache config state not equivalent: tf_data_date changed")
		return false, fmt.Errorf("tf_data_date changed")
	}

	if len(state.TFConfigFileStates) != len(otherState.TFConfigFileStates) {
		logging.Logger.Debug().Msgf("Plan cache config state not equivalent: TFConfigFileStates has changed size")
		return false, fmt.Errorf("tf_config_file_states changed size")
	}

	for i := range state.TFConfigFileStates {
		if state.TFConfigFileStates[i] != otherState.TFConfigFileStates[i] {
			logging.Logger.Debug().Msgf("Plan cache config state not equivalent: %v", state.TFConfigFileStates[i])
			return false, fmt.Errorf("tf_config_file_states changed")
		}
	}

	return true, nil
}

type cacheFile struct {
	ConfigState configState `json:"config_state"`
	Plan        []byte      `json:"plan"`
}

func UsePlanCache(p *DirProvider) bool {
	if p.ctx.RunContext.Config.NoCache {
		// cache was turned off with --no-cache
		return false
	}

	if p.IsTerragrunt {
		// not sure how to support terragrunt yet
		return false
	}

	if p.ctx.RunContext.IsCIRun() {
		return false
	}

	if _, ok := p.ctx.ContextValues.GetValue("terraformRemoteExecutionModeEnabled"); ok {
		// remote execution is enabled
		return false
	}

	return true
}

func ReadPlanCache(p *DirProvider) ([]byte, error) {
	cache := path.Join(calcCacheDir(p), cacheFileName)

	info, err := os.Stat(cache)
	if err != nil {
		logging.Logger.Debug().Msgf("Skipping plan cache: Cache file does not exist")
		p.ctx.CacheErr = "not found"
		return nil, fmt.Errorf("not found")
	}

	if time.Now().Unix()-info.ModTime().Unix() > cacheMaxAgeSecs {
		logging.Logger.Debug().Msgf("Skipping plan cache: Cache file is too old")
		p.ctx.CacheErr = "expired"
		return nil, fmt.Errorf("expired")
	}

	data, err := os.ReadFile(cache)
	if err != nil {
		logging.Logger.Debug().Msgf("Skipping plan cache: Error reading cache file: %v", err)
		p.ctx.CacheErr = "unreadable"
		return nil, fmt.Errorf("unreadable")
	}

	var cf cacheFile
	err = json.Unmarshal(data, &cf)
	if err != nil {
		logging.Logger.Debug().Msgf("Skipping plan cache: Error unmarshalling cache file: %v", err)
		p.ctx.CacheErr = "bad format"
		return nil, fmt.Errorf("bad format")
	}

	state := calcConfigState(p)
	if _, err := cf.ConfigState.equivalent(&state); err != nil {
		logging.Logger.Debug().Msgf("Skipping plan cache: Config state has changed")
		p.ctx.CacheErr = err.Error()
		return nil, fmt.Errorf("change detected")
	}

	logging.Logger.Debug().Msgf("Read plan JSON from %v", cacheFileName)
	p.ctx.UsingCache = true
	return cf.Plan, nil
}

func WritePlanCache(p *DirProvider, planJSON []byte) {
	cacheJSON, err := json.Marshal(cacheFile{ConfigState: calcConfigState(p), Plan: planJSON})
	if err != nil {
		logging.Logger.Debug().Msgf("Failed to marshal plan cache: %v", err)
		return
	}

	cacheDir := calcCacheDir(p)
	// create the .infracost dir if it doesn't already exist
	if _, err := os.Stat(cacheDir); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(cacheDir, 0700)
			if err != nil {
				logging.Logger.Debug().Msgf("Couldn't create %v directory: %v", config.InfracostDir, err)
				return
			}
		}
	}

	err = os.WriteFile(path.Join(cacheDir, cacheFileName), cacheJSON, 0600)
	if err != nil {
		logging.Logger.Debug().Msgf("Failed to write plan cache: %v", err)
		return
	}
	logging.Logger.Debug().Msgf("Wrote plan JSON to %v", cacheFileName)
}

func calcDataDir(p *DirProvider) string {
	if dir, ok := p.Env["TF_DATA_DIR"]; ok {
		return dir
	}

	if dir, ok := os.LookupEnv("TF_DATA_DIR"); ok {
		return dir
	}

	return path.Join(p.Path, ".terraform")
}

func calcCacheDir(p *DirProvider) string {
	dataDir := calcDataDir(p)

	if dataDir != (path.Join(p.Path, ".terraform")) {
		// there is a custom data dir, store the cache under that
		return path.Join(dataDir, config.InfracostDir)
	}

	return path.Join(p.Path, config.InfracostDir)
}

func calcConfigState(p *DirProvider) configState {
	var tfLockFileDate string
	if lockStat, err := os.Stat(path.Join(p.Path, ".terraform.lock.hcl")); err == nil {
		tfLockFileDate = lockStat.ModTime().String()
	}

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
		TFLockFileDate:      tfLockFileDate,
		TFDataDate:          calcTFDataDate(calcDataDir(p), 2).String(),
		TFConfigFileStates:  calcTerraformConfigFileStates(p.Path),
	}
}

func calcTFDataDate(path string, maxDepth int) time.Time {
	var t time.Time

	entries, err := os.ReadDir(path)
	if err == nil {
		for _, entry := range entries {
			if entry.Name() == config.InfracostDir {
				// ignore the infradir since we expect that to change
				continue
			}

			if info, err := entry.Info(); err == nil {
				if t.Before(info.ModTime()) {
					t = info.ModTime()
				}
			}

			if entry.IsDir() {
				if maxDepth > 0 {
					dirT := calcTFDataDate(filepath.Join(path, entry.Name()), maxDepth-1)
					if t.Before(dirT) {
						t = dirT
					}
				}

			}
		}
	}

	return t
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
