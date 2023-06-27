package terraform

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	tgcli "github.com/gruntwork-io/terragrunt/cli"
	"github.com/gruntwork-io/terragrunt/cli/tfsource"
	tgconfig "github.com/gruntwork-io/terragrunt/config"
	tgconfigstack "github.com/gruntwork-io/terragrunt/configstack"
	tgerrors "github.com/gruntwork-io/terragrunt/errors"
	"github.com/gruntwork-io/terragrunt/options"
	tgoptions "github.com/gruntwork-io/terragrunt/options"
	"github.com/gruntwork-io/terragrunt/util"
	"github.com/hashicorp/go-getter"
	hcl2 "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/otiai10/copy"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/gocty"
	ctyJson "github.com/zclconf/go-cty/cty/json"

	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/hcl"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
)

const terragruntSourceVersionFile = ".terragrunt-source-version"

type panicError struct {
	msg string
}

func (p panicError) Error() string {
	return p.msg
}

type TerragruntHCLProvider struct {
	ctx                  *config.ProjectContext
	Path                 string
	includePastResources bool
	outputs              map[string]cty.Value
	stack                *tgconfigstack.Stack
	excludedPaths        []string
	env                  map[string]string
	sourceCache          map[string]string
	logger               *log.Entry
}

// NewTerragruntHCLProvider creates a new provider intialized with the configured project path (usually the terragrunt
// root directory).
func NewTerragruntHCLProvider(ctx *config.ProjectContext, includePastResources bool) schema.Provider {
	logger := ctx.Logger().WithFields(log.Fields{
		"provider": "terragrunt_dir",
	})

	return &TerragruntHCLProvider{
		ctx:                  ctx,
		Path:                 ctx.ProjectConfig.Path,
		includePastResources: includePastResources,
		outputs:              map[string]cty.Value{},
		excludedPaths:        ctx.ProjectConfig.ExcludePaths,
		env:                  getEnvVars(),
		sourceCache:          map[string]string{},
		logger:               logger,
	}
}

func getEnvVars() map[string]string {
	environment := os.Environ()
	environmentMap := make(map[string]string)

	var filterSafe bool
	safe := make(map[string]struct{})
	if v, ok := os.LookupEnv("INFRACOST_SAFE_ENVS"); ok {
		filterSafe = true

		keys := strings.Split(v, ",")
		for _, key := range keys {
			safe[strings.ToLower(strings.TrimSpace(key))] = struct{}{}
		}
	}

	for i := 0; i < len(environment); i++ {
		variableSplit := strings.SplitN(environment[i], "=", 2)

		if len(variableSplit) != 2 {
			continue
		}

		name := strings.TrimSpace(variableSplit[0])
		if !filterSafe {
			environmentMap[name] = variableSplit[1]
		}

		if _, ok := safe[strings.ToLower(name)]; ok {
			environmentMap[name] = variableSplit[1]
		}
	}

	return environmentMap
}

func (p *TerragruntHCLProvider) Type() string {
	return "terragrunt_dir"
}

func (p *TerragruntHCLProvider) DisplayType() string {
	return "Terragrunt directory"
}

func (p *TerragruntHCLProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	metadata.ConfigSha = p.ctx.ProjectConfig.ConfigSha

	basePath := p.ctx.ProjectConfig.Path
	if p.ctx.RunContext.Config.ConfigFilePath != "" {
		basePath = filepath.Dir(p.ctx.RunContext.Config.ConfigFilePath)
	}

	modulePath, err := filepath.Rel(basePath, metadata.Path)
	if err == nil && modulePath != "" && modulePath != "." {
		p.logger.Debugf("Calculated relative terraformModulePath for %s from %s", basePath, metadata.Path)
		metadata.TerraformModulePath = modulePath
	}

	metadata.TerraformWorkspace = p.ctx.ProjectConfig.TerraformWorkspace
}

type terragruntWorkingDirInfo struct {
	configDir  string
	workingDir string
	provider   *HCLProvider
	error      error
	warnings   []schema.ProjectDiag
}

func (i *terragruntWorkingDirInfo) addWarning(pd schema.ProjectDiag) {
	i.warnings = append(i.warnings, pd)
}

// LoadResources finds any Terragrunt projects, prepares them by downloading any required source files, then
// process each with an HCLProvider.
func (p *TerragruntHCLProvider) LoadResources(usage schema.UsageMap) ([]*schema.Project, error) {
	dirs, err := p.prepWorkingDirs()
	if err != nil {
		return nil, err
	}

	var allProjects []*schema.Project

	runCtx := p.ctx.RunContext
	parallelism, _ := runCtx.GetParallelism()

	numJobs := len(dirs)
	runInParallel := parallelism > 1 && numJobs > 1
	if runInParallel && !runCtx.Config.IsLogging() {
		p.logger.Logger.SetLevel(log.InfoLevel)
		p.ctx.RunContext.Config.LogLevel = "info"
	}

	if numJobs < parallelism {
		parallelism = numJobs
	}

	ch := make(chan *terragruntWorkingDirInfo, numJobs)
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for _, dir := range dirs {
		ch <- dir
	}
	close(ch)
	wg.Add(parallelism)

	for i := 0; i < parallelism; i++ {
		go func() {
			defer func() {
				wg.Done()
			}()

			for di := range ch {
				if di.error != nil {
					mu.Lock()
					allProjects = append(allProjects, p.newErroredProject(di))
					mu.Unlock()
					continue
				}

				p.logger.Debugf("Found terragrunt HCL working dir: %v", di.workingDir)

				// HCLProvider.LoadResources never returns an error.
				projects, _ := di.provider.LoadResources(usage)

				for _, project := range projects {
					projectPath := di.configDir
					// attempt to convert project path to be relative to the top level provider path
					if absPath, err := filepath.Abs(p.ctx.ProjectConfig.Path); err == nil {
						if relProjectPath, err := filepath.Rel(absPath, projectPath); err == nil {
							projectPath = filepath.Join(p.ctx.ProjectConfig.Path, relProjectPath)
						}
					}

					metadata := p.newProjectMetadata(projectPath)
					metadata.Warnings = di.warnings
					project.Metadata = metadata
					project.Name = p.generateProjectName(metadata)
					mu.Lock()
					allProjects = append(allProjects, project)
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()
	sort.Slice(allProjects, func(i, j int) bool {
		return allProjects[i].Metadata.TerraformModulePath < allProjects[j].Metadata.TerraformModulePath
	})

	return allProjects, nil
}

func (p *TerragruntHCLProvider) newErroredProject(di *terragruntWorkingDirInfo) *schema.Project {
	projectPath := di.configDir
	if absPath, err := filepath.Abs(p.ctx.ProjectConfig.Path); err == nil {
		if relProjectPath, err := filepath.Rel(absPath, projectPath); err == nil {
			projectPath = filepath.Join(p.ctx.ProjectConfig.Path, relProjectPath)
		}
	}

	metadata := p.newProjectMetadata(projectPath)

	if di.error != nil {
		metadata.AddErrorWithCode(di.error, schema.DiagTerragruntEvaluationFailure)
	}

	return schema.NewProject(p.generateProjectName(metadata), metadata)
}

func (p *TerragruntHCLProvider) generateProjectName(metadata *schema.ProjectMetadata) string {
	name := p.ctx.ProjectConfig.Name
	if name == "" {
		name = metadata.GenerateProjectName(p.ctx.RunContext.VCSMetadata.Remote, p.ctx.RunContext.IsCloudEnabled())
	}
	return name
}

func (p *TerragruntHCLProvider) newProjectMetadata(projectPath string) *schema.ProjectMetadata {
	metadata := config.DetectProjectMetadata(projectPath)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)

	return metadata
}

func (p *TerragruntHCLProvider) initTerraformVarFiles(tfVarFiles []string, extraArgs []tgconfig.TerraformExtraArguments, basePath string) []string {
	v := tfVarFiles

	for _, extraArg := range extraArgs {
		varFiles := extraArg.GetVarFiles(p.logger)
		for _, f := range varFiles {
			absBasePath, _ := filepath.Abs(basePath)
			relPath, err := filepath.Rel(absBasePath, f)
			if err != nil {
				p.logger.Debugf("Error processing var-file, could not get relative path for %s from %s", f, basePath)
				continue
			}

			v = append(v, relPath)
		}
	}

	return v
}

func (p *TerragruntHCLProvider) initTerraformVars(tfVars map[string]string, inputs map[string]interface{}) map[string]string {
	m := make(map[string]string, len(tfVars)+len(inputs))
	for k, v := range tfVars {
		m[k] = v
	}
	for k, v := range inputs {
		m[k] = fmt.Sprintf("%v", v)
	}
	return m
}

func (p *TerragruntHCLProvider) prepWorkingDirs() ([]*terragruntWorkingDirInfo, error) {
	terragruntConfigPath := tgconfig.GetDefaultConfigPath(p.Path)

	terragruntCacheDir := filepath.Join(config.InfracostDir, ".terragrunt-cache")
	terragruntDownloadDir := filepath.Join(p.ctx.RunContext.Config.CachePath(), terragruntCacheDir)
	err := os.MkdirAll(terragruntDownloadDir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("Failed to create download directories for terragrunt in working directory: %w", err)
	}

	mu := sync.Mutex{}
	var workingDirsToEstimate []*terragruntWorkingDirInfo

	tgLog := p.logger.WithFields(log.Fields{"library": "terragrunt"})
	terragruntOptions := &tgoptions.TerragruntOptions{
		TerragruntConfigPath:       terragruntConfigPath,
		Logger:                     tgLog,
		LogLevel:                   log.DebugLevel,
		ErrWriter:                  tgLog.WriterLevel(log.DebugLevel),
		MaxFoldersToCheck:          tgoptions.DEFAULT_MAX_FOLDERS_TO_CHECK,
		WorkingDir:                 p.Path,
		ExcludeDirs:                p.excludedPaths,
		DownloadDir:                terragruntDownloadDir,
		TerraformCliArgs:           []string{tgcli.CMD_TERRAGRUNT_INFO},
		Env:                        p.env,
		IgnoreExternalDependencies: true,
		SourceMap:                  p.ctx.RunContext.Config.TerraformSourceMap,
		RunTerragrunt: func(opts *tgoptions.TerragruntOptions) (err error) {
			defer func() {
				unexpectedErr := recover()
				if unexpectedErr != nil {
					err = panicError{msg: fmt.Sprintf("%s\n%s", unexpectedErr, debug.Stack())}
					mu.Lock()
					workingDirsToEstimate = append(
						workingDirsToEstimate,
						&terragruntWorkingDirInfo{configDir: opts.WorkingDir, workingDir: opts.WorkingDir, error: err},
					)
					mu.Unlock()
				}
			}()

			workingDirInfo := p.runTerragrunt(opts)
			if workingDirInfo != nil {
				mu.Lock()
				workingDirsToEstimate = append(workingDirsToEstimate, workingDirInfo)
				mu.Unlock()
			}

			return
		},
		Functions: func(baseDir string) map[string]function.Function {
			funcs := hcl.ExpFunctions(baseDir, p.logger)

			funcs["run_cmd"] = mockSliceFuncStaticReturn(cty.StringVal("mock-run_cmd"))
			funcs["sops_decrypt_file"] = mockSliceFuncStaticReturn(cty.StringVal("mock"))

			return funcs
		},
		Parallelism: 1,
	}

	terragruntConfigFiles, err := tgconfig.FindConfigFilesInPath(terragruntOptions.WorkingDir, terragruntOptions)
	if err != nil {
		return nil, err
	}

	var filtered []string
	for _, file := range terragruntConfigFiles {
		if !p.isParentTerragruntConfig(file, terragruntConfigFiles) {
			filtered = append(filtered, file)
		}
	}

	// Filter these config files against the exclude paths so Terragrunt doesn't even try to evaluate them
	terragruntConfigFiles = p.filterExcludedPaths(filtered)

	howThesePathsWereFound := fmt.Sprintf("Terragrunt config file found in a subdirectory of %s", terragruntOptions.WorkingDir)
	s, err := createStackForTerragruntConfigPaths(terragruntOptions.WorkingDir, terragruntConfigFiles, terragruntOptions, howThesePathsWereFound)
	if err != nil {
		return nil, err
	}
	p.stack = s

	err = s.Run(terragruntOptions)
	if err != nil {
		return nil, clierror.NewCLIError(
			errors.Errorf(
				"%s\n%v%s",
				"Failed to parse the Terragrunt code using the Terragrunt library:",
				err.Error(),
				fmt.Sprintf("For a list of known issues and workarounds, see: %s", ui.LinkString("https://infracost.io/docs/features/terragrunt/")),
			),
			fmt.Sprintf("Error parsing the Terragrunt code using the Terragrunt library: %s", err),
		)
	}
	p.outputs = map[string]cty.Value{}

	return workingDirsToEstimate, nil
}

// isParentTerragruntConfig checks if a terragrunt config entry is a parent file that is referenced by another config
// with a find_in_parent_folders call. The find_in_parent_folders function searches up the directory tree
// from the file and returns the absolute path to the first terragrunt.hcl. This means if it is found
// we can treat this file as a child terragrunt.hcl.
func (p *TerragruntHCLProvider) isParentTerragruntConfig(parent string, configFiles []string) bool {
	for _, name := range configFiles {
		if !isChildDirectory(parent, name) {
			continue
		}

		file, err := os.Open(name)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			// skip any commented out lines
			if strings.HasPrefix(line, "#") {
				continue
			}

			if strings.Contains(line, "find_in_parent_folders()") {
				file.Close()
				return true
			}
		}

		file.Close()
	}

	return false
}

func isChildDirectory(parent, child string) bool {
	if parent == child {
		return false
	}

	parentDir := filepath.Dir(parent)
	childDir := filepath.Dir(child)
	p, err := filepath.Rel(parentDir, childDir)
	if err != nil || strings.Contains(p, "..") {
		return false
	}

	return true
}

func (p *TerragruntHCLProvider) filterExcludedPaths(paths []string) []string {
	isSkipped := buildExcludedPathsMatcher(p.Path, p.excludedPaths)

	var filteredPaths []string

	for _, path := range paths {
		if !isSkipped(path) {
			filteredPaths = append(filteredPaths, path)
		} else {
			p.logger.Debugf("skipping path %s as it is marked as excluded by --exclude-path", path)
		}
	}

	return filteredPaths
}

// runTerragrunt evaluates a Terragrunt directory with the given opts. This method is called from the
// Terragrunt internal libs as part of the Terraform project evaluation. runTerragrunt is called with for all of the folders
// in a Terragrunt project. Folders that have outputs that are used by other folders are evaluated first.
//
// runTerragrunt will
//  1. build a valid Terraform run env from the opts provided.
//  2. download source modules that are required for the project.
//  3. we then evaluate the Terraform project built by Terragrunt storing any outputs so that we can use
//     these for further runTerragrunt calls that use the dependency outputs.
func (p *TerragruntHCLProvider) runTerragrunt(opts *tgoptions.TerragruntOptions) (info *terragruntWorkingDirInfo) {
	info = &terragruntWorkingDirInfo{configDir: opts.WorkingDir, workingDir: opts.WorkingDir}
	outputs := p.fetchDependencyOutputs(opts)
	terragruntConfig, err := tgconfig.ParseConfigFile(opts.TerragruntConfigPath, opts, nil, &outputs)
	if err != nil {
		info.error = err
		return
	}

	terragruntOptionsClone := opts.Clone(opts.TerragruntConfigPath)
	terragruntOptionsClone.TerraformCommand = tgcli.CMD_TERRAGRUNT_READ_CONFIG

	if terragruntConfig.Skip {
		opts.Logger.Infof(
			"Skipping terragrunt module %s due to skip = true.",
			opts.TerragruntConfigPath,
		)

		return nil
	}

	// get the default download dir
	_, defaultDownloadDir, err := tgoptions.DefaultWorkingAndDownloadDirs(opts.TerragruntConfigPath)
	if err != nil {
		info.error = err
		return
	}

	// if the download dir hasn't been changed from default, and is set in the config,
	// then use it
	if opts.DownloadDir == defaultDownloadDir && terragruntConfig.DownloadDir != "" {
		opts.DownloadDir = terragruntConfig.DownloadDir
	}

	// Override the default value of retryable errors using the value set in the config file
	if terragruntConfig.RetryableErrors != nil {
		opts.RetryableErrors = terragruntConfig.RetryableErrors
	}

	if terragruntConfig.RetryMaxAttempts != nil {
		if *terragruntConfig.RetryMaxAttempts < 1 {
			info.error = fmt.Errorf("Cannot have less than 1 max retry, but you specified %d", *terragruntConfig.RetryMaxAttempts)
			return
		}
		opts.RetryMaxAttempts = *terragruntConfig.RetryMaxAttempts
	}

	if terragruntConfig.RetrySleepIntervalSec != nil {
		if *terragruntConfig.RetrySleepIntervalSec < 0 {
			info.error = fmt.Errorf("Cannot sleep for less than 0 seconds, but you specified %d", *terragruntConfig.RetrySleepIntervalSec)
			return
		}
		opts.RetrySleepIntervalSec = time.Duration(*terragruntConfig.RetrySleepIntervalSec) * time.Second
	}

	sourceURL, err := tgconfig.GetTerraformSourceUrl(opts, terragruntConfig)
	if err != nil {
		info.error = err
		return
	}
	if sourceURL != "" {
		updatedTerragruntOptions, err := p.downloadTerraformSource(sourceURL, opts, terragruntConfig)
		if err != nil {
			info.error = err
			return
		}

		if updatedTerragruntOptions != nil && updatedTerragruntOptions.WorkingDir != "" {
			info = &terragruntWorkingDirInfo{configDir: opts.WorkingDir, workingDir: updatedTerragruntOptions.WorkingDir}
		}
	}

	pconfig := *p.ctx.ProjectConfig // clone the projectConfig
	pconfig.Path = info.workingDir

	if terragruntConfig.Terraform != nil {
		pconfig.TerraformVarFiles = p.initTerraformVarFiles(pconfig.TerraformVarFiles, terragruntConfig.Terraform.ExtraArgs, pconfig.Path)
	}
	pconfig.TerraformVars = p.initTerraformVars(pconfig.TerraformVars, terragruntConfig.Inputs)

	ops := []hcl.Option{
		hcl.OptionWithSpinner(p.ctx.RunContext.NewSpinner),
	}
	inputs, err := convertToCtyWithJson(terragruntConfig.Inputs)
	if err != nil {
		p.logger.Debugf("Failed to build Terragrunt inputs for: %s err: %s", info.workingDir, err)
	} else {
		ops = append(ops, hcl.OptionWithRawCtyInput(inputs))
	}

	fields := p.logger.Data
	fields["parent_provider"] = "terragrunt_dir"

	h, err := NewHCLProvider(
		config.NewProjectContext(p.ctx.RunContext, &pconfig, fields),
		&HCLProviderConfig{CacheParsingModules: true},
		ops...,
	)
	if err != nil {
		projectPath := info.configDir
		if absPath, err := filepath.Abs(p.ctx.ProjectConfig.Path); err == nil {
			if relProjectPath, err := filepath.Rel(absPath, projectPath); err == nil {
				projectPath = filepath.Join(p.ctx.ProjectConfig.Path, relProjectPath)
			}
		}

		info.error = fmt.Errorf("failed to evaluated Terraform directory %s: %w", projectPath, err)
		return
	}

	mods := h.Modules()
	for _, mod := range mods {
		if mod.Error != nil {
			path := ""
			if mod.Module != nil {
				path = mod.Module.RootPath
			}
			info.addWarning(schema.ProjectDiag{
				Code:    schema.DiagTerragruntModuleEvaluationFailure,
				Message: mod.Error.Error(),
			})
			p.logger.Warnf("Terragrunt config path %s returned module %s with error: %s", opts.TerragruntConfigPath, path, mod.Error)
		}
		evaluatedOutputs := mod.Module.Blocks.Outputs(true)
		p.outputs[opts.TerragruntConfigPath] = evaluatedOutputs
	}

	info.provider = h
	return info
}

func buildExcludedPathsMatcher(fullPath string, excludedDirs []string) func(string) bool {
	var excludedMatches []string

	for _, dir := range excludedDirs {
		var absoluteDir string

		if filepath.IsAbs(dir) {
			absoluteDir = dir
		} else {
			absoluteDir, _ = filepath.Abs(filepath.Join(fullPath, dir))
		}

		globs, err := filepath.Glob(absoluteDir)
		if err == nil {
			excludedMatches = append(excludedMatches, globs...)
		}
	}

	return func(dir string) bool {
		absoluteDir, _ := filepath.Abs(dir)

		for _, match := range excludedMatches {
			if strings.HasPrefix(absoluteDir, match) {
				return true
			}
		}

		return false
	}
}

func convertToCtyWithJson(val interface{}) (cty.Value, error) {
	jsonBytes, err := json.Marshal(val)
	if err != nil {
		return cty.DynamicVal, fmt.Errorf("could not marshal terragrunt inputs %w", err)
	}
	var ctyJsonVal ctyJson.SimpleJSONValue
	if err := ctyJsonVal.UnmarshalJSON(jsonBytes); err != nil {
		return cty.DynamicVal, fmt.Errorf("could not unmarshall terragrunt inputs %w", err)
	}
	return ctyJsonVal.Value, nil
}

var (
	depRegexp   = regexp.MustCompile(`dependency\.[\w\-.\[\]"]+`)
	indexRegexp = regexp.MustCompile(`(\w+)\[(\d+)]`)
	mapRegexp   = regexp.MustCompile(`\["([\w\d]+)"]`)
)

func (p *TerragruntHCLProvider) fetchDependencyOutputs(opts *tgoptions.TerragruntOptions) cty.Value {
	moduleOutputs, err := p.fetchModuleOutputs(opts)
	if err != nil {
		p.logger.WithError(err).Debug("failed to fetch real module outputs, defaulting to mocked outputs from file regexp")
	}

	file, err := os.Open(opts.TerragruntConfigPath)
	if err != nil {
		p.logger.WithError(err).Debug("could not open Terragrunt file for dependency regexps")
		return moduleOutputs
	}

	var matches []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// skip any commented out lines
		if strings.HasPrefix(line, "#") {
			continue
		}

		match := depRegexp.FindString(line)
		if match != "" {
			matches = append(matches, match)
		}
	}

	if err := scanner.Err(); err != nil {
		p.logger.WithError(err).Debug("error scanning Terragrunt file lines matching whole file with regexp")

		b, err := os.ReadFile(opts.TerragruntConfigPath)
		if err != nil {
			p.logger.WithError(err).Debug("could not read Terragrunt file for dependency regxps")
		}

		matches = depRegexp.FindAllString(string(b), -1)
	}

	if len(matches) == 0 {
		return moduleOutputs
	}

	valueMap := moduleOutputs.AsValueMap()

	for _, match := range matches {
		pieces := strings.Split(match, ".")
		valueMap = mergeObjectWithDependencyMap(valueMap, pieces[1:])
	}

	return cty.ObjectVal(valueMap)
}

func mergeObjectWithDependencyMap(valueMap map[string]cty.Value, pieces []string) map[string]cty.Value {
	if valueMap == nil {
		valueMap = make(map[string]cty.Value)
	}

	if len(pieces) == 0 {
		return valueMap
	}

	key := pieces[0]
	indexKeys := indexRegexp.FindStringSubmatch(key)
	if len(indexKeys) != 0 {
		index, _ := strconv.Atoi(indexKeys[2])
		return mergeListWithDependencyMap(valueMap, pieces, indexKeys[1], index)
	}

	mapKeys := mapRegexp.FindAllStringSubmatch(key, -1)
	if len(mapKeys) != 0 {
		keys := []string{pieces[0]}
		for _, match := range mapKeys {
			keys = append(keys, match[1])
		}

		split := strings.Split(key, "[")
		key = split[0]
		pieces = append(keys, pieces[1:]...)
	} else {
		key = strings.TrimSuffix(key, "]")
	}

	if len(pieces) == 1 {
		if v, ok := valueMap[key]; ok && v.IsKnown() {
			return valueMap
		}

		valueMap[key] = cty.StringVal(fmt.Sprintf("%s-mock", key))
		return valueMap
	}

	if v, ok := valueMap[key]; ok {
		if v.CanIterateElements() {
			if isList(v) {
				index, _ := strconv.Atoi(pieces[1])

				return mergeListWithDependencyMap(valueMap, pieces[1:], key, index)
			}

			valueMap[key] = cty.ObjectVal(mergeObjectWithDependencyMap(v.AsValueMap(), pieces[1:]))
			return valueMap
		}

		valueMap[key] = cty.ObjectVal(mergeObjectWithDependencyMap(make(map[string]cty.Value), pieces[1:]))
		return valueMap
	}

	valueMap[key] = cty.ObjectVal(mergeObjectWithDependencyMap(make(map[string]cty.Value), pieces[1:]))
	return valueMap
}

func isList(v cty.Value) bool {
	return v.Type().IsTupleType() || v.Type().IsListType()
}

func mergeListWithDependencyMap(valueMap map[string]cty.Value, pieces []string, key string, index int) map[string]cty.Value {
	indexVal := cty.NumberIntVal(int64(index))

	if len(pieces) == 1 {
		if v, ok := valueMap[key]; ok && isList(v) {
			// if we have the index already in the dependency output, and it is known use the existing value.
			// If the value is unknown we need to override it wil a mock as Terragrunt will explode when they
			// try and marshal the cty values to JSON.
			if v.HasIndex(indexVal).True() && v.Index(indexVal).IsKnown() {
				return valueMap
			}

			existing := v.AsValueSlice()
			vals := make([]cty.Value, index+1)
			for i, value := range existing {
				if value.IsKnown() {
					vals[i] = value
					continue
				}

				vals[i] = cty.StringVal(fmt.Sprintf("%s-%d-mock", key, i))
			}

			for i := len(existing); i <= index; i++ {
				vals[i] = cty.StringVal(fmt.Sprintf("%s-%d-mock", key, i))
			}

			valueMap[key] = cty.TupleVal(vals)
			return valueMap
		}

		vals := make([]cty.Value, index+1)
		for i := 0; i <= index; i++ {
			vals[i] = cty.StringVal(fmt.Sprintf("%s-%d-mock", key, i))
		}

		valueMap[key] = cty.ListVal(vals)
		return valueMap
	}

	mockValue := cty.ObjectVal(mergeObjectWithDependencyMap(map[string]cty.Value{}, pieces[1:]))

	if v, ok := valueMap[key]; ok && isList(v) {
		// if we have the index already in the dependency output, and it is known use the existing value.
		// If the value is unknown we need to override it wil a mock as Terragrunt will explode when they
		// try and marshal the cty values to JSON.
		if v.HasIndex(indexVal).True() && v.Index(indexVal).IsKnown() {
			return valueMap
		}

		existing := v.AsValueSlice()
		vals := make([]cty.Value, index+1)
		for i, value := range existing {
			if value.IsKnown() {
				vals[i] = value
				continue
			}

			vals[i] = mockValue
		}

		for i := len(existing); i <= index; i++ {
			vals[i] = mockValue
		}

		valueMap[key] = cty.TupleVal(vals)
		return valueMap
	}

	vals := make([]cty.Value, index+1)
	for i := 0; i <= index; i++ {
		vals[i] = mockValue
	}

	valueMap[key] = cty.ListVal(vals)
	return valueMap
}

// fetchModuleOutputs returns the Terraform outputs from the dependencies of Terragrunt file provided in the opts input.
func (p *TerragruntHCLProvider) fetchModuleOutputs(opts *tgoptions.TerragruntOptions) (cty.Value, error) {
	outputs := cty.MapVal(map[string]cty.Value{
		"outputs": cty.ObjectVal(map[string]cty.Value{
			"mock": cty.StringVal("val"),
		}),
	})

	if p.stack != nil {
		var mod *tgconfigstack.TerraformModule
		for _, module := range p.stack.Modules {
			if module.TerragruntOptions.TerragruntConfigPath == opts.TerragruntConfigPath {
				mod = module
				break
			}
		}

		if mod != nil && len(mod.Dependencies) > 0 {
			blocks, err := decodeDependencyBlocks(mod.TerragruntOptions.TerragruntConfigPath, opts, nil, nil)
			if err != nil {
				return cty.Value{}, fmt.Errorf("could not parse dependency blocks for Terragrunt file %s %w", mod.TerragruntOptions.TerragruntConfigPath, err)
			}

			out := map[string]cty.Value{}
			for dir, dep := range blocks {
				value, evaluated := p.outputs[dir]
				if !evaluated {
					info := p.runTerragrunt(opts.Clone(dir))
					if info != nil && info.error != nil {
						return outputs, fmt.Errorf("could not evaluate dependency %s at dir %s err: %w", dep.Name, dir, err)
					}

					value = p.outputs[dir]
				}

				out[dep.Name] = cty.MapVal(map[string]cty.Value{
					"outputs": value,
				})
			}

			if len(out) > 0 {
				encoded, err := toCtyValue(out, generateTypeFromValuesMap(out))
				if err == nil {
					return encoded, nil
				}

				p.logger.WithError(err).Warn("could not transform output blocks to cty type, using dummy output type")
			}
		}
	}

	return outputs, nil
}

func toCtyValue(val map[string]cty.Value, ty cty.Type) (v cty.Value, err error) {
	defer func() {
		if e := recover(); e != nil {
			trace := debug.Stack()
			v = cty.DynamicVal
			err = fmt.Errorf("recovered from cty panic converting value (%+v) to type (%s) err: %s trace: %s", val, ty.GoString(), e, trace)
		}
	}()

	return gocty.ToCtyValue(val, ty)
}

// generateTypeFromValuesMap takes a values map and returns an object type that has the same number of fields, but
// bound to each type of the underlying evaluated expression. This is the only way the HCL decoder will be happy, as
// object type is the only map type that allows different types for each attribute (cty.Map requires all attributes to
// have the same type.
func generateTypeFromValuesMap(valMap map[string]cty.Value) cty.Type {
	outType := map[string]cty.Type{}
	for k, v := range valMap {
		outType[k] = v.Type()
	}
	return cty.Object(outType)
}

// 1. Download the given source URL, which should use Terraform's module source syntax, into a temporary folder
// 2. Check if module directory exists in temporary folder
// 3. Copy the contents of terragruntOptions.WorkingDir into the temporary folder.
// 4. Set terragruntOptions.WorkingDir to the temporary folder.
//
// See the NewTerraformSource method for how we determine the temporary folder so we can reuse it across multiple
// runs of Terragrunt to avoid downloading everything from scratch every time.
// Copied from github.com/gruntwork-io/terragrunt
func (p *TerragruntHCLProvider) downloadTerraformSource(source string, terragruntOptions *tgoptions.TerragruntOptions, terragruntConfig *tgconfig.TerragruntConfig) (*tgoptions.TerragruntOptions, error) {
	terraformSource, err := tfsource.NewTerraformSource(source, terragruntOptions.DownloadDir, terragruntOptions.WorkingDir, terragruntOptions.Logger)
	if err != nil {
		return nil, err
	}

	if err := p.downloadTerraformSourceIfNecessary(terraformSource, terragruntOptions, terragruntConfig); err != nil {
		return nil, err
	}

	if _, ok := p.sourceCache[terraformSource.CanonicalSourceURL.String()]; !ok {
		terragruntOptions.Logger.Debugf("Adding %s to the source cache", terraformSource.CanonicalSourceURL.String())
		p.sourceCache[terraformSource.CanonicalSourceURL.String()] = terraformSource.DownloadDir
	}

	terragruntOptions.Logger.Debugf("Copying files from %s into %s", terragruntOptions.WorkingDir, terraformSource.WorkingDir)
	var includeInCopy []string
	if terragruntConfig.Terraform != nil && terragruntConfig.Terraform.IncludeInCopy != nil {
		includeInCopy = *terragruntConfig.Terraform.IncludeInCopy
	}
	if err := util.CopyFolderContents(terragruntOptions.WorkingDir, terraformSource.WorkingDir, tgcli.MODULE_MANIFEST_NAME, includeInCopy); err != nil {
		return nil, err
	}

	updatedTerragruntOptions := terragruntOptions.Clone(terragruntOptions.TerragruntConfigPath)

	terragruntOptions.Logger.Debugf("Setting working directory to %s", terraformSource.WorkingDir)
	updatedTerragruntOptions.WorkingDir = terraformSource.WorkingDir

	return updatedTerragruntOptions, nil
}

// copyLocalSource copies the contents of a previously downloaded source folder into the destination folder
func (p *TerragruntHCLProvider) copyLocalSource(prevDest string, dest string, terragruntOptions *tgoptions.TerragruntOptions) error {
	err := os.MkdirAll(dest, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", dest, err)
	}

	// Skip dotfiles and, but keep:
	// 1. Terraform lock files - these are normally committed to source control
	// 2. .terragrunt-source-version files - these are used to determine if the source has changed
	// 3. .infracost dir - this contains any cached third party modules. We can remove this when we move this directory to the root path
	opt := copy.Options{
		Skip: func(src string) (bool, error) {
			base := filepath.Base(src)
			if base == util.TerraformLockFile || base == terragruntSourceVersionFile || base == config.InfracostDir {
				return false, nil
			}

			return strings.HasPrefix(base, "."), nil
		},
		OnSymlink: func(src string) copy.SymlinkAction {
			return copy.Shallow
		},
	}

	err = copy.Copy(prevDest, dest, opt)
	if err != nil {
		return fmt.Errorf("failed to copy source from '%s' to '%s': %w", prevDest, dest, err)
	}

	return nil
}

// Download the specified TerraformSource if the latest code hasn't already been downloaded.
// Copied from github.com/gruntwork-io/terragrunt
func (p *TerragruntHCLProvider) downloadTerraformSourceIfNecessary(terraformSource *tfsource.TerraformSource, terragruntOptions *tgoptions.TerragruntOptions, terragruntConfig *tgconfig.TerragruntConfig) error {
	alreadyLatest, err := p.alreadyHaveLatestCode(terraformSource, terragruntOptions)
	if err != nil {
		return err
	}

	if alreadyLatest {
		if err := p.validateWorkingDir(terraformSource); err != nil {
			return err
		}
		terragruntOptions.Logger.Debugf("Terraform files in %s are up to date. Will not download again.", terraformSource.WorkingDir)
		return nil
	}

	var previousVersion = ""
	// read previous source version
	// https://github.com/gruntwork-io/terragrunt/issues/1921
	if util.FileExists(terraformSource.VersionFile) {
		previousVersion, err = p.readVersionFile(terraformSource)
		if err != nil {
			return err
		}
	}

	// Check if the directory has already been downloaded during this run and is in the source cache
	// If so, we can just copy the files from the previous download to avoid downloading again
	if prevDownloadDir, ok := p.sourceCache[terraformSource.CanonicalSourceURL.String()]; ok {
		terragruntOptions.Logger.Debugf("Source files have already been downloading. Copying files from %s into %s", prevDownloadDir, terraformSource.DownloadDir)
		err := p.copyLocalSource(prevDownloadDir, terraformSource.DownloadDir, terragruntOptions)
		if err != nil {
			terragruntOptions.Logger.Debugf("Failed to copy local source from %s to %s: %v. Will try to redownload", prevDownloadDir, terraformSource.DownloadDir, err)
		} else {
			terragruntOptions.Logger.Debugf("Successfully copied files from %s to %s. Will not download again", prevDownloadDir, terraformSource.DownloadDir)
			return nil
		}
	}

	// When downloading source, we need to process any hooks waiting on `init-from-module`. Therefore, we clone the
	// options struct, set the command to the value the hooks are expecting, and run the download action surrounded by
	// before and after hooks (if any).
	// terragruntOptionsForDownload := terragruntOptions.Clone(terragruntOptions.TerragruntConfigPath)
	// terragruntOptionsForDownload.TerraformCommand = tgcli.CMD_INIT_FROM_MODULE
	// downloadErr := runActionWithHooks("download source", terragruntOptionsForDownload, terragruntConfig, func() error {
	//	return downloadSource(terraformSource, terragruntOptions, terragruntConfig)
	// })
	downloadErr := p.downloadSource(terraformSource, terragruntOptions, terragruntConfig)

	if downloadErr != nil {
		return downloadErr
	}

	if err := terraformSource.WriteVersionFile(); err != nil {
		return err
	}

	if err := p.validateWorkingDir(terraformSource); err != nil {
		return err
	}

	currentVersion, err := terraformSource.EncodeSourceVersion()
	if err != nil {
		return fmt.Errorf("could not encode source version: %w", err)
	}
	// if source versions are different, create file to run init
	// https://github.com/gruntwork-io/terragrunt/issues/1921
	if previousVersion != currentVersion {
		initFile := util.JoinPath(terraformSource.WorkingDir, ".terragrunt-init-required")
		f, createErr := os.Create(initFile)
		if createErr != nil {
			return createErr
		}
		defer f.Close()
	}

	return nil
}

// Download the code from the Canonical Source URL into the Download Folder using the go-getter library
// Copied from github.com/gruntwork-io/terragrunt
func (p *TerragruntHCLProvider) downloadSource(terraformSource *tfsource.TerraformSource, terragruntOptions *tgoptions.TerragruntOptions, terragruntConfig *tgconfig.TerragruntConfig) error {
	terragruntOptions.Logger.Debugf("Downloading Terraform configurations from %s into %s", terraformSource.CanonicalSourceURL, terraformSource.DownloadDir)

	if err := getter.GetAny(terraformSource.DownloadDir, terraformSource.CanonicalSourceURL.String(), p.updateGetters(terragruntConfig)); err != nil {
		return tgerrors.WithStackTrace(err)
	}

	return nil
}

// updateGetters returns the customized go-getter interfaces that Terragrunt relies on. Specifically:
//   - Local file path getter is updated to copy the files instead of creating symlinks, which is what go-getter defaults
//     to.
//   - Include the customized getter for fetching sources from the Terraform Registry.
//
// This creates a closure that returns a function so that we have access to the terragrunt configuration, which is
// necessary for customizing the behavior of the file getter.
// Copied from github.com/gruntwork-io/terragrunt
func (p *TerragruntHCLProvider) updateGetters(terragruntConfig *tgconfig.TerragruntConfig) func(*getter.Client) error {
	return func(client *getter.Client) error {
		// We copy all the default getters from the go-getter library, but replace the "file" getter. We shallow clone the
		// getter map here rather than using getter.Getters directly because (a) we shouldn't change the original,
		// globally-shared getter.Getters map and (b) Terragrunt may run this code from many goroutines concurrently during
		// xxx-all calls, so creating a new map each time ensures we don't a "concurrent map writes" error.
		client.Getters = map[string]getter.Getter{}
		for getterName, getterValue := range getter.Getters {
			if getterName == "file" {
				var includeInCopy []string
				if terragruntConfig.Terraform != nil && terragruntConfig.Terraform.IncludeInCopy != nil {
					includeInCopy = *terragruntConfig.Terraform.IncludeInCopy
				}
				client.Getters[getterName] = &tgcli.FileCopyGetter{IncludeInCopy: includeInCopy}
			} else {
				client.Getters[getterName] = getterValue
			}
		}

		// Load in custom getters that are only supported in Terragrunt
		client.Getters["tfr"] = &TerraformRegistryGetter{}

		return nil
	}
}

// Check if working terraformSource.WorkingDir exists and is directory
// Copied from github.com/gruntwork-io/terragrunt
func (p *TerragruntHCLProvider) validateWorkingDir(terraformSource *tfsource.TerraformSource) error {
	workingLocalDir := strings.ReplaceAll(terraformSource.WorkingDir, terraformSource.DownloadDir+filepath.FromSlash("/"), "")
	if util.IsFile(terraformSource.WorkingDir) {
		return tgcli.WorkingDirNotDir{Dir: workingLocalDir, Source: terraformSource.CanonicalSourceURL.String()}
	}
	if !util.IsDir(terraformSource.WorkingDir) {
		return tgcli.WorkingDirNotFound{Dir: workingLocalDir, Source: terraformSource.CanonicalSourceURL.String()}
	}

	return nil
}

// Returns true if the specified TerraformSource, of the exact same version, has already been downloaded into the
// DownloadFolder. This helps avoid downloading the same code multiple times. Note that if the TerraformSource points
// to a local file path, we assume the user is doing local development and always return false to ensure the latest
// code is downloaded (or rather, copied) every single time. See the ProcessTerraformSource method for more info.
// Copied from github.com/gruntwork-io/terragrunt
func (p *TerragruntHCLProvider) alreadyHaveLatestCode(terraformSource *tfsource.TerraformSource, terragruntOptions *tgoptions.TerragruntOptions) (bool, error) {
	if tfsource.IsLocalSource(terraformSource.CanonicalSourceURL) ||
		!util.FileExists(terraformSource.DownloadDir) ||
		!util.FileExists(terraformSource.WorkingDir) ||
		!util.FileExists(terraformSource.VersionFile) {

		return false, nil
	}

	tfFiles, err := filepath.Glob(fmt.Sprintf("%s/*.tf", terraformSource.WorkingDir))
	if err != nil {
		return false, tgerrors.WithStackTrace(err)
	}

	if len(tfFiles) == 0 {
		terragruntOptions.Logger.Debugf("Working dir %s exists but contains no Terraform files, so assuming code needs to be downloaded again.", terraformSource.WorkingDir)
		return false, nil
	}

	currentVersion, err := terraformSource.EncodeSourceVersion()
	if err != nil {
		return false, fmt.Errorf("could not encode source version: %w", err)
	}

	previousVersion, err := p.readVersionFile(terraformSource)
	if err != nil {
		return false, err
	}

	return previousVersion == currentVersion, nil
}

// Return the version number stored in the DownloadDir. This version number can be used to check if the Terraform code
// that has already been downloaded is the same as the version the user is currently requesting. The version number is
// calculated using the encodeSourceVersion method.
// Copied from github.com/gruntwork-io/terragrunt
func (p *TerragruntHCLProvider) readVersionFile(terraformSource *tfsource.TerraformSource) (string, error) {
	return util.ReadFileAsString(terraformSource.VersionFile)
}

type terragruntDependency struct {
	Dependencies []tgconfig.Dependency `hcl:"dependency,block"`
	Remain       hcl2.Body             `hcl:",remain"`
}

// Find all the Terraform modules in the folders that contain the given Terragrunt config files and assemble those
// modules into a Stack object
func createStackForTerragruntConfigPaths(path string, terragruntConfigPaths []string, terragruntOptions *options.TerragruntOptions, howThesePathsWereFound string) (*tgconfigstack.Stack, error) {
	if len(terragruntConfigPaths) == 0 {
		return nil, tgerrors.WithStackTrace(tgconfigstack.NoTerraformModulesFound)
	}

	modules, err := tgconfigstack.ResolveTerraformModules(terragruntConfigPaths, terragruntOptions, nil, howThesePathsWereFound)
	if err != nil {
		return nil, err
	}

	stack := &tgconfigstack.Stack{Path: path, Modules: modules}
	if err := stack.CheckForCycles(); err != nil {
		return nil, err
	}

	return stack, nil
}

// decodeDependencyBlocks parses the file at filename and returns a map containing all the hcl blocks with the "dependency" label.
// The map is keyed by the full path of the config_path attribute specified in the dependency block.
func decodeDependencyBlocks(filename string, terragruntOptions *tgoptions.TerragruntOptions, dependencyOutputs *cty.Value, include *tgconfig.IncludeConfig) (map[string]tgconfig.Dependency, error) {
	parser := hclparse.NewParser()

	parseFunc := parser.ParseHCLFile
	if strings.HasSuffix(filename, ".json") {
		parseFunc = parser.ParseJSONFile
	}

	file, diags := parseFunc(filename)
	if diags != nil && diags.HasErrors() {
		return nil, fmt.Errorf("could not parse hcl file %s to decode dependency blocks %w", filename, diags)
	}

	localsAsCty, trackInclude, err := tgconfig.DecodeBaseBlocks(terragruntOptions, parser, file, filename, include, nil)
	if err != nil {
		return nil, fmt.Errorf("could not parse base hcl blocks %w", err)
	}

	contextExtensions := tgconfig.EvalContextExtensions{
		Locals:              localsAsCty,
		TrackInclude:        trackInclude,
		DecodedDependencies: dependencyOutputs,
	}

	evalContext, err := tgconfig.CreateTerragruntEvalContext(filename, terragruntOptions, contextExtensions)
	if err != nil {
		return nil, err
	}

	var deps terragruntDependency
	decodeDiagnostics := gohcl.DecodeBody(file.Body, evalContext, &deps)
	if decodeDiagnostics != nil && decodeDiagnostics.HasErrors() {
		return nil, decodeDiagnostics
	}

	depmap := make(map[string]tgconfig.Dependency)
	keymap := make(map[string]struct{})
	for _, dep := range deps.Dependencies {
		depmap[getCleanedTargetConfigPath(dep.ConfigPath, filename)] = dep
		keymap[dep.Name] = struct{}{}
	}

	if trackInclude != nil {
		for _, includeConfig := range trackInclude.CurrentList {
			includeConfig := includeConfig
			strategy, _ := includeConfig.GetMergeStrategy()
			if strategy != tgconfig.NoMerge {
				rawPath := getCleanedTargetConfigPath(includeConfig.Path, filename)
				incl, err := decodeDependencyBlocks(rawPath, terragruntOptions, dependencyOutputs, &includeConfig)
				if err != nil {
					return nil, fmt.Errorf("could not decode dependency blocks for included config '%s' path: %s %w", includeConfig.Name, includeConfig.Path, err)
				}

				for _, dep := range incl {
					if _, includedInParent := keymap[dep.Name]; includedInParent {
						continue
					}

					depmap[getCleanedTargetConfigPath(dep.ConfigPath, filename)] = dep
				}
			}
		}
	}

	return depmap, nil
}

func getCleanedTargetConfigPath(configPath string, workingPath string) string {
	cwd := filepath.Dir(workingPath)
	targetConfig := configPath
	if !filepath.IsAbs(targetConfig) {
		targetConfig = util.JoinPath(cwd, targetConfig)
	}
	if util.IsDir(targetConfig) {
		targetConfig = tgconfig.GetDefaultConfigPath(targetConfig)
	}
	return util.CleanPath(targetConfig)
}

func mockSliceFuncStaticReturn(val cty.Value) function.Function {
	return function.New(&function.Spec{
		VarParam: &function.Parameter{Type: cty.String},
		Type:     function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			return val, nil
		},
	})
}
