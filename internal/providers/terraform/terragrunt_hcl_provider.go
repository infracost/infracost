package terraform

import (
	"bufio"
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
	hcl2 "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
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
	infSync "github.com/infracost/infracost/internal/sync"
	"github.com/infracost/infracost/internal/ui"
)

// terragruntSourceLock is the global lock which works across TerragrunHCLProviders to provide
// concurrency safe downloading.
var terragruntSourceLock = infSync.KeyMutex{}

// terragruntDownloadedDirs is used to ensure sources are only downloaded once. Use terragruntSourceLock
// for concurrency safety.
var terragruntDownloadedDirs = map[string]bool{}

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
		env:                  getEnvVars(ctx),
		sourceCache:          map[string]string{},
		logger:               logger,
	}
}

func getEnvVars(ctx *config.ProjectContext) map[string]string {
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

	for k, v := range ctx.ProjectConfig.Env {
		if !filterSafe {
			environmentMap[k] = v
		}

		if _, ok := safe[strings.ToLower(k)]; ok {
			environmentMap[k] = v
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
		if allProjects[i].Metadata == nil {
			return false
		}

		if allProjects[j].Metadata == nil {
			return true
		}

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
		updatedWorkingDir, err := downloadSourceOnce(sourceURL, opts, terragruntConfig)

		if err != nil {
			info.error = err
			return
		}

		if updatedWorkingDir != "" {
			info = &terragruntWorkingDirInfo{configDir: opts.WorkingDir, workingDir: updatedWorkingDir}
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

// downloadSourceOnce thread-safely makes sure the sourceURL is only downloaded once
func downloadSourceOnce(sourceURL string, opts *tgoptions.TerragruntOptions, terragruntConfig *tgconfig.TerragruntConfig) (string, error) {
	source, err := tfsource.NewTerraformSource(sourceURL, opts.DownloadDir, opts.WorkingDir, opts.Logger)
	if err != nil {
		return "", err
	}

	// make the source download directory absolute so that we lock on a consistent key.
	dir := source.DownloadDir
	if v, err := filepath.Abs(dir); err == nil {
		dir = v
	}

	unlock := terragruntSourceLock.Lock(dir)
	defer unlock()

	if alreadyDownloaded := terragruntDownloadedDirs[dir]; alreadyDownloaded {
		return source.WorkingDir, nil
	}

	_, err = tgcli.DownloadTerraformSource(sourceURL, opts, terragruntConfig)
	if err != nil {
		return "", err
	}

	terragruntDownloadedDirs[dir] = true

	return source.WorkingDir, nil
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
