package providers

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/awslabs/goformation/v7"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/hcl"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/providers/cloudformation"
	"github.com/infracost/infracost/internal/providers/terraform"
	"github.com/infracost/infracost/internal/schema"
)

type DetectionOutput struct {
	Providers   []schema.Provider
	RootModules int
	Tree        string
}

// Detect returns a list of providers for the given path. Multiple returned
// providers are because of auto-detected root modules residing under the
// original path.
func Detect(ctx *config.RunContext, project *config.Project, includePastResources bool) (*DetectionOutput, error) {
	if _, err := os.Stat(project.Path); os.IsNotExist(err) {
		return &DetectionOutput{}, fmt.Errorf("No such file or directory %s", project.Path)
	}

	forceCLI := project.TerraformForceCLI
	projectType := DetectProjectType(project.Path, forceCLI)
	projectContext := config.NewProjectContext(ctx, project, nil)
	if projectType != ProjectTypeAutodetect {
		projectContext.ContextValues.SetValue("project_type", projectType)
	}

	switch projectType {
	case ProjectTypeTerraformPlanJSON:
		return &DetectionOutput{Providers: []schema.Provider{terraform.NewPlanJSONProvider(projectContext, includePastResources)}, RootModules: 1}, nil
	case ProjectTypeTerraformPlanBinary:
		return &DetectionOutput{Providers: []schema.Provider{terraform.NewPlanProvider(projectContext, includePastResources)}, RootModules: 1}, nil
	case ProjectTypeTerraformCLI:
		return &DetectionOutput{Providers: []schema.Provider{terraform.NewDirProvider(projectContext, includePastResources)}, RootModules: 1}, nil
	case ProjectTypeTerragruntCLI:
		return &DetectionOutput{Providers: []schema.Provider{terraform.NewTerragruntProvider(projectContext, includePastResources)}, RootModules: 1}, nil
	case ProjectTypeTerraformStateJSON:
		return &DetectionOutput{Providers: []schema.Provider{terraform.NewStateJSONProvider(projectContext, includePastResources)}, RootModules: 1}, nil
	case ProjectTypeCloudFormation:
		return &DetectionOutput{Providers: []schema.Provider{cloudformation.NewTemplateProvider(projectContext, includePastResources)}, RootModules: 1}, nil
	}

	pathOverrides := make([]hcl.PathOverrideConfig, len(ctx.Config.Autodetect.PathOverrides))
	for i, override := range ctx.Config.Autodetect.PathOverrides {
		pathOverrides[i] = hcl.PathOverrideConfig{
			Path:    override.Path,
			Only:    override.Only,
			Exclude: override.Exclude,
		}
	}

	locatorConfig := &hcl.ProjectLocatorConfig{
		ExcludedDirs:   append(project.ExcludePaths, ctx.Config.Autodetect.ExcludeDirs...),
		IncludedDirs:   ctx.Config.Autodetect.IncludeDirs,
		PathOverrides:  pathOverrides,
		EnvNames:       ctx.Config.Autodetect.EnvNames,
		ChangedObjects: ctx.VCSMetadata.Commit.ChangedObjects,
		UseAllPaths:    project.IncludeAllPaths,
		// If the user has specified terraform var files, we should skip auto-detection
		// as terraform var files are relative to the project root, so invalid path errors
		// will occur if any autodetect projects are outside the current project path.
		SkipAutoDetection:          project.SkipAutodetect || len(project.TerraformVarFiles) > 0,
		FallbackToIncludePaths:     ctx.IsAutoDetect(),
		MaxSearchDepth:             ctx.Config.Autodetect.MaxSearchDepth,
		ForceProjectType:           ctx.Config.Autodetect.ForceProjectType,
		TerraformVarFileExtensions: ctx.Config.Autodetect.TerraformVarFileExtensions,
		PreferFolderNameForEnv:     ctx.Config.Autodetect.PreferFolderNameForEnv,
	}

	// if the config file path is set, we should set the project locator to use the
	// working directory this is so that the paths of the detected RootPaths are
	// relative to the working directory and not the paths specified in the config
	// file.
	if ctx.Config.ConfigFilePath != "" {
		locatorConfig.WorkingDirectory = ctx.Config.WorkingDirectory()
	}

	pl := hcl.NewProjectLocator(logging.Logger, locatorConfig)
	rootPaths, tree := pl.FindRootModules(project.Path)
	if len(rootPaths) == 0 {
		return &DetectionOutput{
			Tree: tree,
		}, fmt.Errorf("could not detect path type for '%s'", project.Path)
	}

	var autoProviders []schema.Provider
	for _, rootPath := range rootPaths {
		detectedProjectContext := config.NewProjectContext(ctx, project, nil)
		if rootPath.IsTerragrunt {
			detectedProjectContext.ContextValues.SetValue("project_type", "terragrunt_dir")
			autoProviders = append(autoProviders, terraform.NewTerragruntHCLProvider(rootPath, detectedProjectContext))
		} else {
			detectedProjectContext.ContextValues.SetValue("project_type", "terraform_dir")
			if ctx.Config.ConfigFilePath == "" && len(project.TerraformVarFiles) == 0 {
				autoProviders = append(autoProviders, autodetectedRootToProviders(detectedProjectContext, rootPath)...)
			} else {
				autoProviders = append(autoProviders, configFileRootToProvider(rootPath, nil, detectedProjectContext, pl))
			}

		}
	}

	return &DetectionOutput{Providers: autoProviders, RootModules: len(rootPaths), Tree: tree}, nil
}

// configFileRootToProvider returns a provider for the given root path which is
// assumed to be a root module defined with a config file. In this case the
// terraform var files should not be grouped/reordered as the user has specified
// these manually.
func configFileRootToProvider(rootPath hcl.RootPath, options []hcl.Option, projectContext *config.ProjectContext, pl *hcl.ProjectLocator) *terraform.HCLProvider {
	var autoVarFiles []string
	for _, varFile := range rootPath.TerraformVarFiles {
		if hcl.IsAutoVarFile(varFile.RelPath) && (filepath.Dir(varFile.RelPath) == rootPath.DetectedPath || filepath.Dir(varFile.RelPath) == ".") {
			autoVarFiles = append(autoVarFiles, varFile.RelPath)
		}
	}

	if len(autoVarFiles) > 0 {
		options = append(options, hcl.OptionWithTFVarsPaths(autoVarFiles, false))
	}

	h, providerErr := terraform.NewHCLProvider(
		projectContext,
		rootPath,
		nil,
		options...,
	)
	if providerErr != nil {
		logging.Logger.Warn().Err(providerErr).Msgf("could not initialize provider for path %q", rootPath.DetectedPath)
	}
	return h
}

// autodetectedRootToProviders returns a list of providers for the given root
// path. These providers are generated by autodetected environments defined in
// the root module. These are defined by var file naming conventions.
func autodetectedRootToProviders(projectContext *config.ProjectContext, rootPath hcl.RootPath, options ...hcl.Option) []schema.Provider {
	var providers []schema.Provider
	autoVarFiles := rootPath.AutoFiles()
	autoVarFiles = append(autoVarFiles, rootPath.GlobalFiles()...)
	varFileGrouping := rootPath.EnvGroupings()

	if len(varFileGrouping) > 0 {
		for _, env := range varFileGrouping {
			provider, err := terraform.NewHCLProvider(
				projectContext,
				rootPath,
				nil,
				append(
					options,
					hcl.OptionWithTFVarsPaths(append(autoVarFiles.ToPaths(), env.TerraformVarFiles.ToPaths()...), true),
					hcl.OptionWithModuleSuffix(rootPath.DetectedPath, env.Name),
				)...)
			if err != nil {
				logging.Logger.Warn().Err(err).Msgf("could not initialize provider for path %q", rootPath.DetectedPath)
				continue
			}

			providers = append(providers, provider)
		}

		return providers
	}

	varFiles := rootPath.EnvFiles()
	providerOptions := options
	if len(autoVarFiles) > 0 {
		providerOptions = append(providerOptions, hcl.OptionWithTFVarsPaths(append(varFiles.ToPaths(), autoVarFiles.ToPaths()...), true))
	}

	provider, err := terraform.NewHCLProvider(
		projectContext,
		rootPath,
		nil,
		providerOptions...,
	)
	if err != nil {
		logging.Logger.Warn().Err(err).Msgf("could not initialize provider for path %q", rootPath.DetectedPath)
		return nil
	}

	return []schema.Provider{provider}
}

type ProjectType string

var (
	ProjectTypeTerraformPlanJSON   ProjectType = "terraform_plan_json"
	ProjectTypeTerraformPlanBinary ProjectType = "terraform_plan_binary"
	ProjectTypeTerraformCLI        ProjectType = "terraform_cli"
	ProjectTypeTerragruntCLI       ProjectType = "terragrunt_cli"
	ProjectTypeTerraformStateJSON  ProjectType = "terraform_state_json"
	ProjectTypeCloudFormation      ProjectType = "cloudformation"
	ProjectTypeAutodetect          ProjectType = "autodetect"
)

func DetectProjectType(path string, forceCLI bool) ProjectType {
	if isCloudFormationTemplate(path) {
		return ProjectTypeCloudFormation
	}

	if isTerraformPlanJSON(path) {
		return ProjectTypeTerraformPlanJSON
	}

	if isTerraformStateJSON(path) {
		return ProjectTypeTerraformStateJSON
	}

	if isTerraformPlan(path) {
		return ProjectTypeTerraformPlanBinary
	}

	if forceCLI {
		if isTerragruntNestedDir(path, 5) {
			return ProjectTypeTerragruntCLI
		}

		return ProjectTypeTerraformCLI
	}

	return ProjectTypeAutodetect
}

func isTerraformPlanJSON(path string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	var jsonFormat struct {
		FormatVersion string `json:"format_version"`
		PlannedValues any    `json:"planned_values"`
	}

	b, hasWrapper := terraform.StripSetupTerraformWrapper(b)
	if hasWrapper {
		logging.Logger.Info().Msgf("Stripped wrapper output from %s (to make it a valid JSON file) since setup-terraform GitHub Action was used without terraform_wrapper: false", path)
	}

	err = json.Unmarshal(b, &jsonFormat)
	if err != nil {
		return false
	}

	return jsonFormat.FormatVersion != "" && jsonFormat.PlannedValues != nil
}

func isTerraformStateJSON(path string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	var jsonFormat struct {
		FormatVersion string `json:"format_version"`
		Values        any    `json:"values"`
	}

	b, hasWrapper := terraform.StripSetupTerraformWrapper(b)
	if hasWrapper {
		logging.Logger.Debug().Msgf("Stripped setup-terraform wrapper output from %s", path)
	}

	err = json.Unmarshal(b, &jsonFormat)
	if err != nil {
		return false
	}

	return jsonFormat.FormatVersion != "" && jsonFormat.Values != nil
}

func isTerraformPlan(path string) bool {
	r, err := zip.OpenReader(path)
	if err != nil {
		return false
	}
	defer r.Close()

	var planFile *zip.File
	for _, file := range r.File {
		if file.Name == "tfplan" {
			planFile = file
			break
		}
	}

	return planFile != nil
}

func isTerragruntDir(path string) bool {
	if val, ok := os.LookupEnv("TERRAGRUNT_CONFIG"); ok {
		if filepath.IsAbs(val) {
			return config.FileExists(val)
		}
		return config.FileExists(filepath.Join(path, val))
	}

	return config.FileExists(filepath.Join(path, "terragrunt.hcl")) || config.FileExists(filepath.Join(path, "terragrunt.hcl.json"))
}

func isTerragruntNestedDir(path string, maxDepth int) bool {
	if isTerragruntDir(path) {
		return true
	}

	if maxDepth > 0 {
		entries, err := os.ReadDir(path)
		if err == nil {
			for _, entry := range entries {
				name := entry.Name()
				if entry.IsDir() && name != config.InfracostDir && name != ".terraform" {
					if isTerragruntNestedDir(filepath.Join(path, name), maxDepth-1) {
						return true
					}
				}
			}
		}
	}
	return false
}

// goformation lib is not threadsafe, so we run this check synchronously
// See: https://github.com/awslabs/goformation/issues/363
var cfMux = &sync.Mutex{}

func isCloudFormationTemplate(path string) bool {
	cfMux.Lock()
	defer cfMux.Unlock()

	template, err := goformation.Open(path)
	if err != nil {
		return false
	}

	if len(template.Resources) > 0 {
		return true
	}

	return false
}
