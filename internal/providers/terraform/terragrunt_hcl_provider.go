package terraform

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	goerrors "github.com/go-errors/errors"
	tgerrors "github.com/gruntwork-io/go-commons/errors"
	tgcliterraform "github.com/gruntwork-io/terragrunt/cli/commands/terraform"
	tgcliinfo "github.com/gruntwork-io/terragrunt/cli/commands/terragrunt-info"
	"github.com/gruntwork-io/terragrunt/codegen"
	tgconfig "github.com/gruntwork-io/terragrunt/config"
	tgconfigstack "github.com/gruntwork-io/terragrunt/configstack"
	tgoptions "github.com/gruntwork-io/terragrunt/options"
	tgterraform "github.com/gruntwork-io/terragrunt/terraform"
	"github.com/gruntwork-io/terragrunt/util"
	"github.com/hashicorp/go-getter"
	hcl2 "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/gocty"
	ctyJson "github.com/zclconf/go-cty/cty/json"

	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/hcl"
	"github.com/infracost/infracost/internal/hcl/mock"
	"github.com/infracost/infracost/internal/hcl/modules"
	"github.com/infracost/infracost/internal/metrics"
	"github.com/infracost/infracost/internal/schema"
	infSync "github.com/infracost/infracost/internal/sync"
	"github.com/infracost/infracost/internal/ui"
)

var (
	// terragruntSourceLock is the global lock which works across TerragrunHCLProviders to provide
	// concurrency safe downloading.
	terragruntSourceLock = infSync.KeyMutex{}

	// terragruntWorkingDirLock is the global lock which works across TerragrunHCLProviders to provide
	// concurrency safe downloading.
	terragruntWorkingDirLock = infSync.KeyMutex{}

	// terragruntDownloadedDirs is used to ensure sources are only downloaded once. This is needed
	// because the call to util.CopyFolderContents in tgcliterraform.DownloadTerraformSource seems to be mucking
	// up the cache directory unnecessarily.  If that is fixed we can remove this.
	terragruntDownloadedDirs = sync.Map{}

	terragruntOutputCache = &TerragruntOutputCache{
		cache: sync.Map{},
		mu:    &infSync.KeyMutex{},
	}
)

type panicError struct {
	msg string
}

func (p panicError) Error() string {
	return p.msg
}

type TerragruntOutputCache struct {
	cache sync.Map
	mu    *infSync.KeyMutex
}

// Set stores a value in the cache for the given key using the value returned
// from getVal function. If the key already exists in the cache, the value is
// returned from the cache.
func (o *TerragruntOutputCache) Set(key string, getVal func() (cty.Value, error)) (cty.Value, error) {
	unlock := o.mu.Lock(key)
	defer unlock()

	cacheVal, exists := o.cache.Load(key)
	if exists {
		if val, ok := cacheVal.(cty.Value); ok {
			return val, nil
		}
	}

	val, err := getVal()
	if err != nil {
		return cty.NilVal, err
	}

	o.cache.Store(key, val)
	return val, nil
}

type TerragruntHCLProvider struct {
	ctx            *config.ProjectContext
	Path           hcl.RootPath
	stack          *tgconfigstack.Stack
	excludedPaths  []string
	env            map[string]string
	sourceCache    map[string]string
	packageFetcher *modules.PackageFetcher
	logger         zerolog.Logger
}

// NewTerragruntHCLProvider creates a new provider initialized with the configured project path (usually the terragrunt
// root directory).
func NewTerragruntHCLProvider(rootPath hcl.RootPath, ctx *config.ProjectContext) schema.Provider {
	logger := ctx.Logger().With().Str(
		"provider", "terragrunt_dir",
	).Logger()

	var remoteCache modules.RemoteCache
	runCtx := ctx.RunContext
	if runCtx.Config.S3ModuleCacheRegion != "" && runCtx.Config.S3ModuleCacheBucket != "" {
		s3ModuleCache, err := modules.NewS3Cache(runCtx.Config.S3ModuleCacheRegion, runCtx.Config.S3ModuleCacheBucket, runCtx.Config.S3ModuleCachePrefix, runCtx.Config.S3ModuleCachePrivate)
		if err != nil {
			logger.Warn().Msgf("failed to initialize S3 module cache: %s", err)
		} else {
			remoteCache = s3ModuleCache
		}
	}

	fetcher := modules.NewPackageFetcher(remoteCache, logger, modules.WithGetters(map[string]getter.Getter{
		"tfr": &tgterraform.RegistryGetter{
			ProxyForDomains: []string{".terraform.io"},
			ProxyURL:        os.Getenv("INFRACOST_REGISTRY_PROXY"),
		},
		"file": &tgcliterraform.FileCopyGetter{},
	}), modules.WithPublicModuleChecker(modules.NewHttpPublicModuleChecker()))

	return &TerragruntHCLProvider{
		ctx:            ctx,
		Path:           rootPath,
		excludedPaths:  ctx.ProjectConfig.ExcludePaths,
		env:            getEnvVars(ctx),
		sourceCache:    map[string]string{},
		packageFetcher: fetcher,
		logger:         logger,
	}
}

func getEnvVars(ctx *config.ProjectContext) map[string]string {
	environment := os.Environ()
	environmentMap := make(map[string]string)

	var filterSafe bool
	safe := make(map[string]struct{})
	if v, ok := os.LookupEnv("INFRACOST_SAFE_ENVS"); ok {
		filterSafe = true

		keys := strings.SplitSeq(v, ",")
		for key := range keys {
			safe[strings.ToLower(strings.TrimSpace(key))] = struct{}{}
		}
	}

	for i := range environment {
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

func (p *TerragruntHCLProvider) Context() *config.ProjectContext { return p.ctx }

func (p *TerragruntHCLProvider) ProjectName() string {
	if p.ctx.ProjectConfig.Name != "" {
		return p.ctx.ProjectConfig.Name
	}

	return config.CleanProjectName(p.RelativePath())
}

func (p *TerragruntHCLProvider) EnvName() string {
	return ""
}

func (p *TerragruntHCLProvider) RelativePath() string {
	r, err := filepath.Rel(p.Path.StartingPath, p.Path.DetectedPath)
	if err != nil {
		return p.Path.DetectedPath
	}

	return r
}

func (p *TerragruntHCLProvider) VarFiles() []string {
	return nil
}

func (p *TerragruntHCLProvider) DependencyPaths() []string {
	return nil
}

func (p *TerragruntHCLProvider) YAML() string {
	str := strings.Builder{}

	str.WriteString(fmt.Sprintf("  - path: %s\n    name: %s\n", p.RelativePath(), p.ProjectName()))

	return str.String()
}

func (p *TerragruntHCLProvider) Type() string {
	return "terragrunt_dir"
}

func (p *TerragruntHCLProvider) DisplayType() string {
	return "Terragrunt"
}

func (p *TerragruntHCLProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	metadata.ConfigSha = p.ctx.ProjectConfig.ConfigSha

	modulePath := p.RelativePath()
	if modulePath != "" && modulePath != "." {
		metadata.TerraformModulePath = modulePath
	}

	metadata.TerraformWorkspace = p.ctx.ProjectConfig.TerraformWorkspace
}

type terragruntWorkingDirInfo struct {
	configDir        string
	workingDir       string
	provider         *HCLProvider
	error            error
	warnings         []*schema.ProjectDiag
	evaluatedOutputs cty.Value
}

// LoadResources finds any Terragrunt projects, prepares them by downloading any required source files, then
// process each with an HCLProvider.
func (p *TerragruntHCLProvider) LoadResources(usage schema.UsageMap) ([]*schema.Project, error) {
	loadResourcesTimer := metrics.GetTimer("terragrunt.LoadResources", false, p.ctx.ProjectConfig.Path).Start()
	defer loadResourcesTimer.Stop()

	dirs, err := p.prepWorkingDirs()
	if err != nil {
		return nil, err
	}

	var allProjects []*schema.Project

	runCtx := p.ctx.RunContext
	parallelism, _ := runCtx.GetParallelism()

	numJobs := len(dirs)
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

				p.logger.Debug().Msgf("Found terragrunt HCL working dir: %v", di.workingDir)

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

					metadata := p.newProjectMetadata(projectPath, project.Metadata)
					metadata.Warnings = di.warnings
					project.Metadata = metadata
					project.Name = p.generateProjectName(metadata)
					project.DisplayName = p.ProjectName()
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

	metadata := p.newProjectMetadata(projectPath, nil)

	if di.error != nil {
		metadata.AddError(schema.NewDiagTerragruntEvaluationFailure(di.error))
	}

	project := schema.NewProject(p.generateProjectName(metadata), metadata)
	project.DisplayName = p.ProjectName()

	return project
}

func (p *TerragruntHCLProvider) generateProjectName(metadata *schema.ProjectMetadata) string {
	name := p.ctx.ProjectConfig.Name
	if name == "" {
		name = metadata.GenerateProjectName(p.ctx.RunContext.VCSMetadata.Remote, p.ctx.RunContext.IsCloudEnabled())
	}
	return name
}

func (p *TerragruntHCLProvider) newProjectMetadata(projectPath string, originalMetadata *schema.ProjectMetadata) *schema.ProjectMetadata {
	metadata := schema.DetectProjectMetadata(projectPath)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)
	if originalMetadata != nil {
		metadata.PolicySha = originalMetadata.PolicySha
		metadata.PastPolicySha = originalMetadata.PastPolicySha
	}

	return metadata
}

func (p *TerragruntHCLProvider) initTerraformVarFiles(tfVarFiles []string, extraArgs []tgconfig.TerraformExtraArguments, basePath string, opts *tgoptions.TerragruntOptions) []string {
	v := tfVarFiles

	for _, extraArg := range extraArgs {
		varFiles := extraArg.GetVarFiles(opts.Logger)
		for _, f := range varFiles {
			absBasePath, _ := filepath.Abs(basePath)
			relPath, err := filepath.Rel(absBasePath, f)
			if err != nil {
				p.logger.Debug().Msgf("Error processing var-file, could not get relative path for %s from %s", f, basePath)
				continue
			}

			v = append(v, relPath)
		}
	}

	return v
}

func (p *TerragruntHCLProvider) initTerraformVars(tfVars map[string]any, inputs map[string]any) map[string]any {
	m := make(map[string]any, len(tfVars)+len(inputs))
	maps.Copy(m, tfVars)
	for k, v := range inputs {
		m[k] = fmt.Sprintf("%v", v)
	}
	return m
}

func (p *TerragruntHCLProvider) prepWorkingDirs() ([]*terragruntWorkingDirInfo, error) {
	terragruntConfigPath := tgconfig.GetDefaultConfigPath(p.Path.DetectedPath)

	terragruntCacheDir := filepath.Join(config.InfracostDir, ".terragrunt-cache")
	terragruntDownloadDir := filepath.Join(p.ctx.RunContext.Config.CachePath(), terragruntCacheDir)
	err := os.MkdirAll(terragruntDownloadDir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("Failed to create download directories for terragrunt in working directory: %w", err)
	}

	mu := sync.Mutex{}
	var workingDirsToEstimate []*terragruntWorkingDirInfo

	tgLog := logrus.StandardLogger().WithFields(logrus.Fields{"library": "terragrunt"})
	terragruntOptions := &tgoptions.TerragruntOptions{
		TerragruntConfigPath:       terragruntConfigPath,
		Logger:                     tgLog,
		LogLevel:                   logrus.DebugLevel,
		ErrWriter:                  tgLog.WriterLevel(logrus.DebugLevel),
		MaxFoldersToCheck:          tgoptions.DefaultMaxFoldersToCheck,
		WorkingDir:                 p.Path.DetectedPath,
		ExcludeDirs:                p.excludedPaths,
		DownloadDir:                terragruntDownloadDir,
		TerraformCliArgs:           []string{tgcliinfo.CommandName},
		Env:                        p.env,
		IgnoreExternalDependencies: true,
		SourceMap:                  p.ctx.RunContext.Config.TerraformSourceMap,
		DownloadSource:             p.downloadSource,
		UsePartialParseConfigCache: true,
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
			_, _ = terragruntOutputCache.Set(opts.TerragruntConfigPath, func() (cty.Value, error) {
				if workingDirInfo == nil {
					return cty.EmptyObjectVal, errors.New("nil outputs")
				}

				return workingDirInfo.evaluatedOutputs, nil
			})
			if workingDirInfo != nil {
				mu.Lock()
				workingDirsToEstimate = append(workingDirsToEstimate, workingDirInfo)
				mu.Unlock()
			}

			return
		},
		Functions: func(baseDir string) map[string]function.Function {
			funcs := hcl.ExpFunctions(baseDir, p.logger)

			funcs["run_cmd"] = mockSliceFuncStaticReturn(cty.StringVal(fmt.Sprintf("run_cmd-%s", mock.Identifier)))
			funcs["sops_decrypt_file"] = mockSliceFuncStaticReturn(cty.StringVal(fmt.Sprintf("sops_decrypt_file-%s", mock.Identifier)))
			funcs["get_aws_account_id"] = mockSliceFuncStaticReturn(cty.StringVal(fmt.Sprintf("account_id-%s", mock.Identifier)))
			funcs["get_aws_caller_identity_arn"] = mockSliceFuncStaticReturn(cty.StringVal(fmt.Sprintf("arn:aws:iam::123456789012:user/terragrunt-%s", mock.Identifier)))
			funcs["get_aws_caller_identity_user_id"] = mockSliceFuncStaticReturn(cty.StringVal(fmt.Sprintf("caller_identity_user_id-%s", mock.Identifier)))

			return funcs
		},
		Parallelism: 1,
		GetOutputs:  p.terragruntPathToValue,
		DiagnosticsFunc: func(_ error, filename string, config any, evalContext *hcl2.EvalContext) {
			configFile := config.(*tgconfig.TerragruntConfigFile)
			f, err := hclparse.NewParser().ParseHCLFile(filename)
			if err != nil {
				p.logger.Warn().Msgf("Terragrunt diagnostic func failed to reparse Terragrunt config file %s: %s", filename, err)
				return
			}

			content, _, err := f.Body.PartialContent(&hcl2.BodySchema{
				Attributes: []hcl2.AttributeSchema{
					{
						Name: "inputs",
					},
				},
			})
			if err != nil {
				p.logger.Debug().Msgf("Terragrunt diagnostic func failed to reparse Terragrunt inputs for %s: %s", filename, err)
				return
			}

			attr := hcl.Attribute{
				HCLAttr: content.Attributes["inputs"],
				Ctx:     hcl.NewContext(evalContext, nil, p.logger),
				Logger:  p.logger,
				// set is graph to true so that we use the better expression mocking
				// when the expression hits a diagnostic.
				IsGraph: true,
			}
			v := attr.Value()

			configFile.Inputs = &v
		},
	}

	howThesePathsWereFound := fmt.Sprintf("Terragrunt config file found in a subdirectory of %s", terragruntOptions.WorkingDir)
	s, err := createStackForTerragruntConfigPaths(terragruntOptions.WorkingDir, []string{
		terragruntConfigPath,
	}, terragruntOptions, howThesePathsWereFound)
	if err != nil {
		return nil, err
	}
	p.stack = s

	err = s.Run(terragruntOptions)
	if err != nil {
		return nil, clierror.NewCLIError(
			fmt.Errorf(
				"%s\n%v%s",
				"Failed to parse the Terragrunt code using the Terragrunt library:",
				err.Error(),
				fmt.Sprintf("For a list of known issues and workarounds, see: %s", ui.LinkString("https://infracost.io/docs/features/terragrunt/")),
			),
			fmt.Sprintf("Error parsing the Terragrunt code using the Terragrunt library: %s", err),
		)
	}

	return workingDirsToEstimate, nil
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
		updatedWorkingDir, err := loadSourceOnce(sourceURL, opts, terragruntConfig, p.packageFetcher, p.logger)

		if err != nil {
			info.error = err
			return
		}

		if updatedWorkingDir != "" {
			info = &terragruntWorkingDirInfo{configDir: opts.WorkingDir, workingDir: updatedWorkingDir}
		}
	}

	unlock := terragruntWorkingDirLock.Lock(info.workingDir)
	defer unlock()

	if err = generateConfig(terragruntConfig, opts, info.workingDir); err != nil {
		info.error = err
		return
	}

	pconfig := *p.ctx.ProjectConfig // clone the projectConfig
	pconfig.Path = info.workingDir

	if terragruntConfig.Terraform != nil {
		pconfig.TerraformVarFiles = p.initTerraformVarFiles(pconfig.TerraformVarFiles, terragruntConfig.Terraform.ExtraArgs, pconfig.Path, opts)
	}
	pconfig.TerraformVars = p.initTerraformVars(pconfig.TerraformVars, terragruntConfig.Inputs)

	var ops []hcl.Option
	inputs, err := convertToCtyWithJson(terragruntConfig.Inputs)
	if err != nil {
		p.logger.Debug().Msgf("Failed to build Terragrunt inputs for: %s err: %s", info.workingDir, err)
	} else {
		ops = append(ops, hcl.OptionWithRawCtyInput(inputs))
	}

	logCtx := p.logger.With().Str("parent_provider", "terragrunt_dir").Ctx(context.Background())

	h, err := NewHCLProvider(
		config.NewProjectContext(p.ctx.RunContext, &pconfig, logCtx),
		hcl.RootPath{
			DetectedPath: pconfig.Path,
		},
		&HCLProviderConfig{CacheParsingModules: true, SkipAutoDetection: true},
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

	mod := h.Module()
	info.error = mod.Error
	info.provider = h
	info.evaluatedOutputs = mod.Module.Blocks.Outputs(true)
	return info
}

// downloadSource overrides the Terragrunt download source functionality to use the Infracost packageFetcher
// so that it makes use of any cached modules
func (p *TerragruntHCLProvider) downloadSource(downloadDir string, sourceURL string, opts *tgoptions.TerragruntOptions) error {
	return p.packageFetcher.Fetch(sourceURL, downloadDir)
}

func splitModuleSubDir(moduleSource string) (string, string, error) {
	moduleAddr, submodulePath := getter.SourceDirSubdir(moduleSource)
	if strings.HasPrefix(submodulePath, "../") {
		return "", "", fmt.Errorf("invalid submodule path '%s'", submodulePath)
	}

	return moduleAddr, submodulePath, nil
}

// loadSourceOnce thread-safely makes sure the sourceURL is only downloaded once.
// It calls the internal Terragrunt functionality to download the source.
func loadSourceOnce(sourceURL string, opts *tgoptions.TerragruntOptions, terragruntConfig *tgconfig.TerragruntConfig, packageFetcher *modules.PackageFetcher, logger zerolog.Logger) (string, error) {
	_, modAddr, err := splitModuleSubDir(sourceURL)
	if err != nil {
		return "", err
	}

	parsedSourceURL, err := url.Parse(sourceURL)
	if err != nil {
		return "", err
	}

	sparseCheckoutEnabled := os.Getenv("INFRACOST_SPARSE_CHECKOUT") == "true"

	// If sparse checkout is enabled add the subdir to the Source URL as a query param
	// so go-getter only downloads the required directory.
	if sparseCheckoutEnabled {
		q := parsedSourceURL.Query()
		q.Set("subdir", modAddr)
		parsedSourceURL.RawQuery = q.Encode()
		sourceURL = parsedSourceURL.String()
	}

	source, err := tgterraform.NewSource(sourceURL, opts.DownloadDir, opts.WorkingDir, terragruntConfig.GenerateConfigs, opts.Logger)
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

	if _, alreadyDownloaded := terragruntDownloadedDirs.Load(dir); alreadyDownloaded {
		return source.WorkingDir, nil
	}

	// first attempt to force an HTTPS download of the source, this is only applicable to
	// SSH sources. We do this to try and make use of any git credentials stored on a
	// host machine, rather than requiring an SSH key is added.
	failedHttpsDownload := !forceHttpsDownload(sourceURL, opts, terragruntConfig)
	if failedHttpsDownload {
		_, err = tgcliterraform.DownloadTerraformSource(sourceURL, opts, terragruntConfig)
		if err != nil {
			return "", err
		}
	}

	if sparseCheckoutEnabled && modAddr != "" && isGitDir(dir) {
		symlinkedDirs, err := modules.ResolveSymLinkedDirs(dir, modAddr)
		if err != nil {
			return "", err
		}

		if len(symlinkedDirs) > 0 {
			mu := &sync.Mutex{}
			logger.Trace().Msgf("recursively adding symlinked dirs to sparse-checkout for repo %s: %v", dir, symlinkedDirs)
			// Using a depth of 1 here since the submodule directory is already downloaded, so only need
			// to add the symlinked directories to the sparse-checkout.
			err := modules.RecursivelyAddDirsToSparseCheckout(dir, sourceURL, packageFetcher, []string{modAddr}, symlinkedDirs, mu, logger, 1)
			if err != nil {
				return "", err
			}
		}
	}

	terragruntDownloadedDirs.Store(dir, true)

	return source.WorkingDir, nil
}

// isGitDir checks if the directory is a git directory
func isGitDir(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}

func forceHttpsDownload(sourceURL string, opts *tgoptions.TerragruntOptions, terragruntConfig *tgconfig.TerragruntConfig) bool {
	newSource, err := getter.Detect(sourceURL, opts.WorkingDir, getter.Detectors)
	if err != nil {
		return false
	}
	u, err := url.Parse(newSource)
	if err != nil {
		return false
	}

	if !modules.IsGitSSHSource(u) {
		return false
	}

	newUrl, err := modules.NormalizeGitURLToHTTPS(u)
	if err != nil {
		return false
	}

	_, err = tgcliterraform.DownloadTerraformSource(newUrl.String(), opts, terragruntConfig)
	return err == nil
}

func generateConfig(terragruntConfig *tgconfig.TerragruntConfig, opts *tgoptions.TerragruntOptions, workingDir string) error {
	unlock := terragruntSourceLock.Lock(opts.DownloadDir)
	defer unlock()

	for _, config := range terragruntConfig.GenerateConfigs {
		if err := codegen.WriteToFile(opts, workingDir, config); err != nil {
			return err
		}
	}

	return nil
}

func convertToCtyWithJson(val any) (cty.Value, error) {
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
	// depRegexp is used to find dependency references in Terragrunt files, e.g. dependency.foo.bar
	// so that we can mock the outputs of these dependencies into a "shape" that is expected by any
	// inputs that reference it.
	// The following regex supports formats like:
	// - dependency.foo.bar
	// - dependency.foo.bar[0]
	// - dependency.foo.bar["baz"]
	// - dependency.foo.bar[0]["baz"]
	// - tolist(dependency.foo.bar)
	// - tolist(dependency.foo.bar)[0]
	// - tomap(dependency.foo.bar)["baz"]
	// - tolist(dependency.foo.bar["baz")[0]
	// - tomap(dependency.foo.bar[0])["baz"]
	// - values(dependency.foo.bar).*
	// - values(dependency.foo.bar).*.baz
	// - values(dependency.foo.bar).baz[0]
	// - values(dependency.foo.bar).baz["qux"]
	// - values(dependency.foo.bar).*[0]
	// - values(dependency.foo.bar).*[0]["qux"]
	// - values(dependency.foo.bar).*.baz[0]
	// - values(dependency.foo.bar).*.baz["qux"]
	//
	// The regex will also match any trailing braces so we have to strip them later on.
	// There's no easy way to avoid this since golang regex doesn't support lookbehinds.
	depRegexp   = regexp.MustCompile(`(?:\w+\()?dependency\.[\w\-.\[\]"]+(?:\)[\w\*\-.\[\]"]+)?(?:\))?`)
	indexRegexp = regexp.MustCompile(`(\w+)\[(\d+)]`)
	mapRegexp   = regexp.MustCompile(`\["([\w\d]+)"]`)
)

func (p *TerragruntHCLProvider) fetchDependencyOutputs(opts *tgoptions.TerragruntOptions) cty.Value {
	moduleOutputs, err := p.fetchModuleOutputs(opts)
	if err != nil {
		p.logger.Debug().Err(err).Msg("failed to fetch real module outputs, defaulting to mocked outputs from file regexp")
	}

	file, err := os.Open(opts.TerragruntConfigPath)
	if err != nil {
		p.logger.Debug().Err(err).Msg("could not open Terragrunt file for dependency regexps")
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
		p.logger.Debug().Err(err).Msg("error scanning Terragrunt file lines matching whole file with regexp")

		b, err := os.ReadFile(opts.TerragruntConfigPath)
		if err != nil {
			p.logger.Debug().Err(err).Msg("could not read Terragrunt file for dependency regxps")
		}

		matches = depRegexp.FindAllString(string(b), -1)
	}

	if len(matches) == 0 {
		return moduleOutputs
	}

	valueMap := moduleOutputs.AsValueMap()

	for _, match := range matches {
		stripped := stripTrailingBraces(stripFunctionCalls(match))
		pieces := strings.Split(stripped, ".")
		valueMap = mergeObjectWithDependencyMap(valueMap, pieces[1:])
	}

	return cty.ObjectVal(valueMap)
}

func stripFunctionCalls(input string) string {
	re := regexp.MustCompile(`\w+\(([^)]+)\)`)
	stripped := re.ReplaceAllString(input, "$1")

	return stripped
}

func stripTrailingBraces(input string) string {
	return strings.TrimSuffix(input, ")")
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

		valueMap[key] = cty.StringVal(fmt.Sprintf("%s-%s", key, mock.Identifier))
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

				vals[i] = cty.StringVal(fmt.Sprintf("%s-%d-%s", key, i, mock.Identifier))
			}

			for i := len(existing); i <= index; i++ {
				vals[i] = cty.StringVal(fmt.Sprintf("%s-%d-%s", key, i, mock.Identifier))
			}

			valueMap[key] = cty.TupleVal(vals)
			return valueMap
		}

		vals := make([]cty.Value, index+1)
		for i := 0; i <= index; i++ {
			vals[i] = cty.StringVal(fmt.Sprintf("%s-%d-%s", key, i, mock.Identifier))
		}

		valueMap[key] = cty.ListVal(vals)
		return valueMap
	}

	mockValue := cty.ObjectVal(mergeObjectWithDependencyMap(map[string]cty.Value{}, pieces[1:]))

	if v, ok := valueMap[key]; ok && isList(v) {
		existing := v.AsValueSlice()

		s := max(len(existing), index+1)
		vals := make([]cty.Value, s)

		for i := range existing {
			existingVal := existing[i]
			if i != index && existingVal.IsKnown() {
				vals[i] = existingVal
				continue
			}

			// if we are at the index and the value is known, and it is an object we need to merge the object
			// with mock values for the rest of the pieces.
			if i == index && existingVal.IsKnown() && existingVal.CanIterateElements() && !isList(existingVal) {
				vals[i] = cty.ObjectVal(mergeObjectWithDependencyMap(existingVal.AsValueMap(), pieces[1:]))
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
	fallbackOutputs := cty.MapVal(map[string]cty.Value{
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
			value, err := p.decodeTerragruntDepsToValue(mod.TerragruntOptions.TerragruntConfigPath, opts)
			if err != nil {
				return fallbackOutputs, err
			}

			return value, nil
		}
	}

	return fallbackOutputs, nil
}

func (p *TerragruntHCLProvider) terragruntPathToValue(targetConfig string, opts *tgoptions.TerragruntOptions) (*cty.Value, bool, error) {
	value, err := terragruntOutputCache.Set(targetConfig, func() (cty.Value, error) {
		info := p.runTerragrunt(opts.Clone(targetConfig))
		if info != nil && info.error != nil {
			return cty.EmptyObjectVal, fmt.Errorf("could not run teragrunt path %s err: %w", targetConfig, info.error)
		}

		if info == nil {
			return cty.EmptyObjectVal, nil
		}

		return info.evaluatedOutputs, nil
	})

	if err != nil {
		return &cty.EmptyObjectVal, true, err
	}

	if value.RawEquals(cty.EmptyObjectVal) {
		return &cty.EmptyObjectVal, true, nil
	}

	return &value, false, nil
}

func (p *TerragruntHCLProvider) decodeTerragruntDepsToValue(targetConfig string, opts *tgoptions.TerragruntOptions) (cty.Value, error) {
	blocks, err := decodeDependencyBlocks(targetConfig, opts, nil, nil)
	if err != nil {
		return cty.EmptyObjectVal, fmt.Errorf("could not parse dependency blocks for Terragrunt file %s %w", targetConfig, err)
	}

	out := map[string]cty.Value{}
	for dir, dep := range blocks {
		value, _, depErr := p.terragruntPathToValue(dir, opts)
		if depErr != nil {
			return cty.EmptyObjectVal, depErr
		}

		out[dep.Name] = cty.MapVal(map[string]cty.Value{
			"outputs": *value,
		})
	}

	if len(out) > 0 {
		encoded, err := toCtyValue(out, generateTypeFromValuesMap(out))
		if err == nil {
			return encoded, nil
		}

		p.logger.Debug().Err(err).Msg("could not transform output blocks to cty type, using dummy output type")
	}

	return cty.EmptyObjectVal, nil
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
func createStackForTerragruntConfigPaths(path string, terragruntConfigPaths []string, terragruntOptions *tgoptions.TerragruntOptions, howThesePathsWereFound string) (*tgconfigstack.Stack, error) {
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
	var withStack *goerrors.Error
	if errors.As(err, &withStack) {
		stack := withStack.ErrorStack()
		return nil, fmt.Errorf("could not parse base hcl blocks %w", errors.New(stack))
	}
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
