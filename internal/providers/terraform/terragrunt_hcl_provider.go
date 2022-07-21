package terraform

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/gruntwork-io/terragrunt/aws_helper"
	tgcli "github.com/gruntwork-io/terragrunt/cli"
	"github.com/gruntwork-io/terragrunt/cli/tfsource"
	tgconfig "github.com/gruntwork-io/terragrunt/config"
	tgconfigstack "github.com/gruntwork-io/terragrunt/configstack"
	tgerrors "github.com/gruntwork-io/terragrunt/errors"
	tgoptions "github.com/gruntwork-io/terragrunt/options"
	"github.com/gruntwork-io/terragrunt/util"
	"github.com/hashicorp/go-getter"
	hcl2 "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	ctyJson "github.com/zclconf/go-cty/cty/json"

	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/hcl"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
)

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
		env:                  parseEnvironmentVariables(os.Environ()),
		logger:               logger,
	}
}

func parseEnvironmentVariables(environment []string) map[string]string {
	environmentMap := make(map[string]string)

	for i := 0; i < len(environment); i++ {
		variableSplit := strings.SplitN(environment[i], "=", 2)

		if len(variableSplit) == 2 {
			environmentMap[strings.TrimSpace(variableSplit[0])] = variableSplit[1]
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
}

// LoadResources finds any Terragrunt projects, prepares them by downloading any required source files, then
// process each with an HCLProvider.
func (p *TerragruntHCLProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	dirs, err := p.prepWorkingDirs()
	if err != nil {
		return nil, err
	}

	var allProjects []*schema.Project

	// Sort the dirs so they are consistent in the output
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].configDir < dirs[j].configDir
	})

	for _, di := range dirs {
		p.logger.Debugf("Found terragrunt HCL working dir: %v", di.workingDir)

		projects, err := di.provider.LoadResources(usage)
		if err != nil {
			return nil, err
		}

		for _, project := range projects {
			projectPath := di.configDir
			// attempt to convert project path to be relative to the top level provider path
			if absPath, err := filepath.Abs(p.ctx.ProjectConfig.Path); err == nil {
				if relProjectPath, err := filepath.Rel(absPath, projectPath); err == nil {
					projectPath = filepath.Join(p.ctx.ProjectConfig.Path, relProjectPath)
				}
			}

			project.Metadata = config.DetectProjectMetadata(projectPath)
			project.Metadata.Type = p.Type()
			p.AddMetadata(project.Metadata)
			project.Name = schema.GenerateProjectName(project.Metadata, p.ctx.ProjectConfig.Name, p.ctx.RunContext.IsCloudEnabled())
			allProjects = append(allProjects, project)
		}
	}

	return allProjects, nil
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

	terragruntDownloadDir := filepath.Join(p.Path, ".infracost/.terragrunt-cache")
	err := os.MkdirAll(terragruntDownloadDir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("Failed to create download directories for terragrunt in working directory: %w", err)
	}

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
		RunTerragrunt: func(terragruntOptions *tgoptions.TerragruntOptions) (err error) {
			defer func() {
				unexpectedErr := recover()
				if unexpectedErr != nil {
					err = panicError{msg: fmt.Sprintf("%s\n%s", unexpectedErr, debug.Stack())}
				}
			}()

			workingDirInfo, err := p.runTerragrunt(terragruntOptions)
			if workingDirInfo != nil {
				workingDirsToEstimate = append(workingDirsToEstimate, workingDirInfo)
			}

			return
		},
		Parallelism: 1,
	}

	s, err := tgconfigstack.FindStackInSubfolders(terragruntOptions)
	if err != nil {
		return nil, err
	}
	p.stack = s

	err = s.Run(terragruntOptions)
	if err != nil {
		if errors.As(err, &panicError{}) {
			panic(err)
		}

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

// runTerragrunt evaluates a Terragrunt directory with the given opts. This method is called from the
// Terragrunt internal libs as part of the Terraform project evaluation. runTerragrunt is called with for all of the folders
// in a Terragrunt project. Folders that have outputs that are used by other folders are evaluated first.
//
// runTerragrunt will
//		1. build a valid Terraform run env from the opts provided.
//		2. download source modules that are required for the project.
// 		3. we then evaluate the Terraform project built by Terragrunt storing any outputs so that we can use
//			these for further runTerragrunt calls that use the dependency outputs.
func (p *TerragruntHCLProvider) runTerragrunt(opts *tgoptions.TerragruntOptions) (*terragruntWorkingDirInfo, error) {
	outputs, err := p.fetchDependencyOutputs(opts)
	if err != nil {
		return nil, err
	}

	terragruntConfig, err := tgconfig.ParseConfigFile(opts.TerragruntConfigPath, opts, nil, &outputs)
	if err != nil {
		return nil, err
	}

	terragruntOptionsClone := opts.Clone(opts.TerragruntConfigPath)
	terragruntOptionsClone.TerraformCommand = tgcli.CMD_TERRAGRUNT_READ_CONFIG

	if terragruntConfig.Skip {
		opts.Logger.Infof(
			"Skipping terragrunt module %s due to skip = true.",
			opts.TerragruntConfigPath,
		)
		return nil, nil
	}

	// We merge the OriginalIAMRoleOptions into the one from the config, because the CLI passed IAMRoleOptions has
	// precedence.
	opts.IAMRoleOptions = tgoptions.MergeIAMRoleOptions(
		terragruntConfig.GetIAMRoleOptions(),
		opts.OriginalIAMRoleOptions,
	)

	if err := aws_helper.AssumeRoleAndUpdateEnvIfNecessary(opts); err != nil {
		return nil, err
	}

	// get the default download dir
	_, defaultDownloadDir, err := tgoptions.DefaultWorkingAndDownloadDirs(opts.TerragruntConfigPath)
	if err != nil {
		return nil, err
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
			return nil, fmt.Errorf("Cannot have less than 1 max retry, but you specified %d", *terragruntConfig.RetryMaxAttempts)
		}
		opts.RetryMaxAttempts = *terragruntConfig.RetryMaxAttempts
	}

	if terragruntConfig.RetrySleepIntervalSec != nil {
		if *terragruntConfig.RetrySleepIntervalSec < 0 {
			return nil, fmt.Errorf("Cannot sleep for less than 0 seconds, but you specified %d", *terragruntConfig.RetrySleepIntervalSec)
		}
		opts.RetrySleepIntervalSec = time.Duration(*terragruntConfig.RetrySleepIntervalSec) * time.Second
	}

	info := &terragruntWorkingDirInfo{configDir: opts.WorkingDir, workingDir: opts.WorkingDir}

	sourceURL, err := tgconfig.GetTerraformSourceUrl(opts, terragruntConfig)
	if err != nil {
		return nil, err
	}
	if sourceURL != "" {
		updatedTerragruntOptions, err := p.downloadTerraformSource(sourceURL, opts, terragruntConfig)
		if err != nil {
			return nil, err
		}
		if updatedTerragruntOptions != nil && updatedTerragruntOptions.WorkingDir != "" {
			info = &terragruntWorkingDirInfo{configDir: opts.WorkingDir, workingDir: updatedTerragruntOptions.WorkingDir}
		}
	}

	pconfig := *p.ctx.ProjectConfig // clone the projectConfig
	pconfig.Path = info.workingDir
	pconfig.TerraformVars = p.initTerraformVars(pconfig.TerraformVars, terragruntConfig.Inputs)

	inputs, err := convertToCtyWithJson(terragruntConfig.Inputs)
	if err != nil {
		p.logger.Debugf("Failed to build Terragrunt inputs for: %s err: %s", info.workingDir, err)
	}

	fields := p.logger.Data
	fields["parent_provider"] = "terragrunt_dir"

	h, err := NewHCLProvider(
		config.NewProjectContext(p.ctx.RunContext, &pconfig, fields),
		&HCLProviderConfig{CacheParsingModules: true},
		hcl.OptionWithSpinner(p.ctx.RunContext.NewSpinner),
		hcl.OptionWithWarningFunc(p.ctx.RunContext.NewWarningWriter()),
		hcl.OptionWithRawCtyInput(inputs),
	)
	if err != nil {
		return nil, fmt.Errorf("could not create provider for Terragrunt generated dir %w", err)
	}

	mods, err := h.Modules()
	if err != nil {
		return nil, fmt.Errorf("could not parse generated Terraform dir from Terragrunt generated dir %w", err)
	}

	for _, mod := range mods {
		p.outputs[opts.TerragruntConfigPath] = mod.Blocks.Outputs(true)
	}

	info.provider = h
	return info, nil
}

func convertToCtyWithJson(val interface{}) (cty.Value, error) {
	jsonBytes, err := json.Marshal(val)
	if err != nil {
		return cty.NilVal, fmt.Errorf("could not marshal terragrunt inputs %w", err)
	}
	var ctyJsonVal ctyJson.SimpleJSONValue
	if err := ctyJsonVal.UnmarshalJSON(jsonBytes); err != nil {
		return cty.NilVal, fmt.Errorf("could not unmarshall terragrunt inputs %w", err)
	}
	return ctyJsonVal.Value, nil
}

// fetchDependencyOutputs returns the Terraform outputs from the dependencies of Terragrunt file provided in the opts input.
func (p *TerragruntHCLProvider) fetchDependencyOutputs(opts *tgoptions.TerragruntOptions) (cty.Value, error) {
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
					_, err := p.runTerragrunt(opts.Clone(dir))
					if err != nil {
						return outputs, fmt.Errorf("could not evaluate dependency %s at dir %s err: %w", dep.Name, dir, err)
					}

					value = p.outputs[dir]
				}

				out[dep.Name] = cty.MapVal(map[string]cty.Value{
					"outputs": value,
				})
			}

			if len(out) > 0 {
				encoded, err := gocty.ToCtyValue(out, generateTypeFromValuesMap(out))
				if err == nil {
					return encoded, nil
				}

				p.logger.WithError(err).Warn("could not transform output blocks to cty type, using dummy output type")
			}
		}
	}

	return outputs, nil
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

// Download the specified TerraformSource if the latest code hasn't already been downloaded.
// Copied from github.com/gruntwork-io/terragrunt
func (p *TerragruntHCLProvider) downloadTerraformSourceIfNecessary(terraformSource *tfsource.TerraformSource, terragruntOptions *tgoptions.TerragruntOptions, terragruntConfig *tgconfig.TerragruntConfig) error {
	if terragruntOptions.SourceUpdate {
		terragruntOptions.Logger.Debugf("The --%s flag is set, so deleting the temporary folder %s before downloading source.", "terragrunt-source-update", terraformSource.DownloadDir)
		if err := os.RemoveAll(terraformSource.DownloadDir); err != nil {
			return tgerrors.WithStackTrace(err)
		}
	}

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

	currentVersion := terraformSource.EncodeSourceVersion()
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
// - Local file path getter is updated to copy the files instead of creating symlinks, which is what go-getter defaults
//   to.
// - Include the customized getter for fetching sources from the Terraform Registry.
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

	currentVersion := terraformSource.EncodeSourceVersion()
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

// decodeDependencyBlocks parses the file at filename and returns a map containing all the hcl blocks with the "dependency" label.
// The map is keyed by the full path of the config_path attribute specified in the dependency block.
func decodeDependencyBlocks(filename string, terragruntOptions *tgoptions.TerragruntOptions, dependencyOutputs *cty.Value, include *tgconfig.IncludeConfig) (map[string]tgconfig.Dependency, error) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCLFile(filename)
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
