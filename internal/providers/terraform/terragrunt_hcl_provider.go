package terraform

import (
	"fmt"
	"github.com/gruntwork-io/terragrunt/aws_helper"
	tgcli "github.com/gruntwork-io/terragrunt/cli"
	"github.com/gruntwork-io/terragrunt/cli/tfsource"
	tgconfig "github.com/gruntwork-io/terragrunt/config"
	tgconfigstack "github.com/gruntwork-io/terragrunt/configstack"
	tgerrors "github.com/gruntwork-io/terragrunt/errors"
	tgoptions "github.com/gruntwork-io/terragrunt/options"
	"github.com/gruntwork-io/terragrunt/util"
	"github.com/infracost/infracost/internal/hcl"
	"strings"

	"github.com/hashicorp/go-getter"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
)

type TerragruntHCLProvider struct {
	ctx                  *config.ProjectContext
	Path                 string
	includePastResources bool
}

// NewTerragruntHCLProvider creates a new provider intialized with the configured project path (usually the terragrunt
// root directory).
func NewTerragruntHCLProvider(ctx *config.ProjectContext, includePastResources bool) schema.Provider {
	return &TerragruntHCLProvider{
		ctx:                  ctx,
		Path:                 ctx.ProjectConfig.Path,
		includePastResources: includePastResources,
	}
}

func (p *TerragruntHCLProvider) Type() string {
	return "terragrunt_hcl"
}

func (p *TerragruntHCLProvider) DisplayType() string {
	return "Terragrunt directory (HCL)"
}

func (p *TerragruntHCLProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	// no op
}

type terragruntWorkingDirInfo struct {
	ConfigDir  string
	WorkingDir string
	Inputs     map[string]interface{}
}

// LoadResources finds any Terragrunt projects, prepares them by downloading any required source files, then
// process each with an HCLProvider.
func (p *TerragruntHCLProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	workingDirInfos, err := p.prepWorkingDirs()
	if err != nil {
		return nil, err
	}

	var allProjects []*schema.Project

	for _, di := range workingDirInfos {
		log.Debugf("Found terragrunt HCL working dir: %v", di.WorkingDir)

		pconfig := *p.ctx.ProjectConfig // clone the projectConfig
		pconfig.Path = di.WorkingDir
		for k, v := range di.Inputs {
			pconfig.TerraformVars = append(pconfig.TerraformVars, fmt.Sprintf("%s=%v", k, v))
		}

		pctx := config.NewProjectContext(p.ctx.RunContext, &pconfig)
		h, err := NewHCLProvider(
			pctx,
			NewPlanJSONProvider(pctx, p.includePastResources),
			hcl.OptionWithSpinner(p.ctx.RunContext.NewSpinner),
			hcl.OptionWithWarningFunc(p.ctx.RunContext.NewWarningWriter()),
		)

		if err != nil {
			return nil, err
		}

		projects, err := h.LoadResources(usage)
		if err != nil {
			return nil, err
		}

		for _, project := range projects {
			metadata := config.DetectProjectMetadata(di.ConfigDir)
			metadata.Type = p.Type()
			p.AddMetadata(metadata)
			project.Name = schema.GenerateProjectName(metadata, p.ctx.RunContext.Config.EnableDashboard)
			allProjects = append(allProjects, project)
		}
	}

	return allProjects, nil
}

func (p *TerragruntHCLProvider) prepWorkingDirs() ([]*terragruntWorkingDirInfo, error) {
	terragruntConfigPath := tgconfig.GetDefaultConfigPath(p.Path)

	terragruntDownloadDir := filepath.Join(p.Path, ".infracost/.terragrunt-cache")
	err := os.MkdirAll(terragruntDownloadDir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("Failed to create download directories for terragrunt in working directory: %w", err)
	}

	var workingDirsToEstimate []*terragruntWorkingDirInfo

	terragruntOptions := &tgoptions.TerragruntOptions{
		TerragruntConfigPath: terragruntConfigPath,
		Logger:               log.WithField("library", "terragrunt"),
		MaxFoldersToCheck:    tgoptions.DEFAULT_MAX_FOLDERS_TO_CHECK,
		WorkingDir:           p.Path,
		DownloadDir:          terragruntDownloadDir,
		TerraformCliArgs:     []string{tgcli.CMD_TERRAGRUNT_INFO},
		RunTerragrunt: func(terragruntOptions *tgoptions.TerragruntOptions) error {
			workingDirInfo, err := p.runTerragrunt(terragruntOptions)
			if workingDirInfo != nil {
				workingDirsToEstimate = append(workingDirsToEstimate, workingDirInfo)
			}
			return err
		},
		Parallelism: 1,
	}

	s, err := tgconfigstack.FindStackInSubfolders(terragruntOptions)
	if err != nil {
		return nil, err
	}

	err = s.Run(terragruntOptions)
	if err != nil {
		return nil, err
	}

	return workingDirsToEstimate, nil
}

// Downloads terraform source if necessary, then runs terraform with the given options and CLI args.
// This will forward all the args and extra_arguments directly to Terraform.
func (p *TerragruntHCLProvider) runTerragrunt(terragruntOptions *tgoptions.TerragruntOptions) (*terragruntWorkingDirInfo, error) {
	terragruntConfig, err := tgconfig.ReadTerragruntConfig(terragruntOptions)
	if err != nil {
		return nil, err
	}

	terragruntOptionsClone := terragruntOptions.Clone(terragruntOptions.TerragruntConfigPath)
	terragruntOptionsClone.TerraformCommand = tgcli.CMD_TERRAGRUNT_READ_CONFIG

	if terragruntConfig.Skip {
		terragruntOptions.Logger.Infof(
			"Skipping terragrunt module %s due to skip = true.",
			terragruntOptions.TerragruntConfigPath,
		)
		return nil, nil
	}

	// We merge the OriginalIAMRoleOptions into the one from the config, because the CLI passed IAMRoleOptions has
	// precedence.
	terragruntOptions.IAMRoleOptions = tgoptions.MergeIAMRoleOptions(
		terragruntConfig.GetIAMRoleOptions(),
		terragruntOptions.OriginalIAMRoleOptions,
	)

	if err := aws_helper.AssumeRoleAndUpdateEnvIfNecessary(terragruntOptions); err != nil {
		return nil, err
	}

	// get the default download dir
	_, defaultDownloadDir, err := tgoptions.DefaultWorkingAndDownloadDirs(terragruntOptions.TerragruntConfigPath)
	if err != nil {
		return nil, err
	}

	// if the download dir hasn't been changed from default, and is set in the config,
	// then use it
	if terragruntOptions.DownloadDir == defaultDownloadDir && terragruntConfig.DownloadDir != "" {
		terragruntOptions.DownloadDir = terragruntConfig.DownloadDir
	}

	// Override the default value of retryable errors using the value set in the config file
	if terragruntConfig.RetryableErrors != nil {
		terragruntOptions.RetryableErrors = terragruntConfig.RetryableErrors
	}

	if terragruntConfig.RetryMaxAttempts != nil {
		if *terragruntConfig.RetryMaxAttempts < 1 {
			return nil, fmt.Errorf("Cannot have less than 1 max retry, but you specified %d", *terragruntConfig.RetryMaxAttempts)
		}
		terragruntOptions.RetryMaxAttempts = *terragruntConfig.RetryMaxAttempts
	}

	if terragruntConfig.RetrySleepIntervalSec != nil {
		if *terragruntConfig.RetrySleepIntervalSec < 0 {
			return nil, fmt.Errorf("Cannot sleep for less than 0 seconds, but you specified %d", *terragruntConfig.RetrySleepIntervalSec)
		}
		terragruntOptions.RetrySleepIntervalSec = time.Duration(*terragruntConfig.RetrySleepIntervalSec) * time.Second
	}

	sourceURL, err := tgconfig.GetTerraformSourceUrl(terragruntOptions, terragruntConfig)
	if err != nil {
		return nil, err
	}
	if sourceURL != "" {
		updatedTerragruntOptions, err := p.downloadTerraformSource(sourceURL, terragruntOptions, terragruntConfig)
		if err != nil {
			return nil, err
		}
		if updatedTerragruntOptions != nil && updatedTerragruntOptions.WorkingDir != "" {
			return &terragruntWorkingDirInfo{ConfigDir: terragruntOptions.WorkingDir, WorkingDir: updatedTerragruntOptions.WorkingDir, Inputs: terragruntConfig.Inputs}, nil
		}
	}

	return &terragruntWorkingDirInfo{ConfigDir: terragruntOptions.WorkingDir, WorkingDir: terragruntOptions.WorkingDir, Inputs: terragruntConfig.Inputs}, nil
}

// 1. Download the given source URL, which should use Terraform's module source syntax, into a temporary folder
// 2. Check if module directory exists in temporary folder
// 3. Copy the contents of terragruntOptions.WorkingDir into the temporary folder.
// 4. Set terragruntOptions.WorkingDir to the temporary folder.
//
// See the NewTerraformSource method for how we determine the temporary folder so we can reuse it across multiple
// runs of Terragrunt to avoid downloading everything from scratch every time.
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
func (p *TerragruntHCLProvider) readVersionFile(terraformSource *tfsource.TerraformSource) (string, error) {
	return util.ReadFileAsString(terraformSource.VersionFile)
}
