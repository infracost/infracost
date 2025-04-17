package terraform

import (
	"crypto/sha256"
	"encoding/hex"
	stdJson "encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	giturls "github.com/chainguard-dev/git-urls"
	"github.com/infracost/infracost/internal/metrics"
	"github.com/infracost/infracost/internal/schema"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	ctyJson "github.com/zclconf/go-cty/cty/json"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/hcl"
	"github.com/infracost/infracost/internal/hcl/modules"
	"github.com/infracost/infracost/internal/logging"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary

	moduleCacheRegex = regexp.MustCompile(`(?:.+)?\.infracost/terraform_modules/[^/]+/(.+)`)
)

type HCLProvider struct {
	policyClient   *apiclient.PolicyAPIClient
	Parser         *hcl.Parser
	planJSONParser *Parser
	logger         zerolog.Logger

	schema *PlanSchema
	ctx    *config.ProjectContext
	cache  *HCLProject
	config HCLProviderConfig
}

type HCLProviderConfig struct {
	SuppressLogging     bool
	CacheParsingModules bool
	SkipAutoDetection   bool
}

type flagStringSlice []string

func (v *flagStringSlice) String() string { return "" }
func (v *flagStringSlice) Set(raw string) error {
	*v = append(*v, raw)
	return nil
}

type vars struct {
	files []string
	vars  []string
}

var spaceReg = regexp.MustCompile(`\s+`)

func varsFromPlanFlags(planFlags string) (vars, error) {
	f := flag.NewFlagSet("", flag.ContinueOnError)

	var fs flagStringSlice
	var vs flagStringSlice

	f.Var(&vs, "var", "")
	f.Var(&fs, "var-file", "")
	err := f.Parse(spaceReg.Split(planFlags, -1))
	if err != nil {
		return vars{}, err
	}

	return vars{
		files: fs,
		vars:  vs,
	}, nil
}

// NewHCLProvider returns a HCLProvider with a hcl.Parser initialised using the config.ProjectContext.
// It will use input flags from either the terraform-plan-flags or top level var and var-file flags to
// set input vars and files on the underlying hcl.Parser.
func NewHCLProvider(ctx *config.ProjectContext, rootPath hcl.RootPath, config *HCLProviderConfig, opts ...hcl.Option) (*HCLProvider, error) {
	if config == nil {
		config = &HCLProviderConfig{}
	}

	v, err := varsFromPlanFlags(ctx.ProjectConfig.TerraformPlanFlags)
	if err != nil {
		return nil, fmt.Errorf("could not parse vars from plan flags %w", err)
	}

	options := []hcl.Option{hcl.OptionWithTFEnvVars(ctx.ProjectConfig.Env), hcl.OptionWithSpaceliftRemoteVarLoader(ctx)}

	if len(v.vars) > 0 {
		withPlanFlagVars := hcl.OptionWithPlanFlagVars(v.vars)
		options = append(options, withPlanFlagVars)
	}

	v.files = append(v.files, ctx.ProjectConfig.TerraformVarFiles...)
	if len(v.files) > 0 {
		withFiles := hcl.OptionWithTFVarsPaths(v.files, false)
		options = append(options, withFiles)
	}

	if len(ctx.ProjectConfig.TerraformVars) > 0 {
		withInputVars := hcl.OptionWithInputVars(ctx.ProjectConfig.TerraformVars)
		options = append(options, withInputVars)
	}

	options = append(options, opts...)

	credsSource, err := modules.NewTerraformCredentialsSource(modules.BaseCredentialSet{
		Token: ctx.ProjectConfig.TerraformCloudToken,
		Host:  ctx.ProjectConfig.TerraformCloudHost,
	})
	localWorkspace := ctx.ProjectConfig.TerraformWorkspace
	if err == nil {
		var loaderOpts []hcl.TFCRemoteVariablesLoaderOption
		if ctx.ProjectConfig.TerraformCloudWorkspace != "" && ctx.ProjectConfig.TerraformCloudOrg != "" {
			loaderOpts = append(loaderOpts, hcl.RemoteVariablesLoaderWithRemoteConfig(hcl.TFCRemoteConfig{
				Organization: ctx.ProjectConfig.TerraformCloudOrg,
				Workspace:    ctx.ProjectConfig.TerraformCloudWorkspace,
				Host:         credsSource.BaseCredentialSet.Host,
			}))
		}

		options = append(options, hcl.OptionWithTFCRemoteVarLoader(
			credsSource.BaseCredentialSet.Host,
			credsSource.BaseCredentialSet.Token,
			localWorkspace,
			loaderOpts...),
		)
	}

	options = append(options,
		hcl.OptionWithTerraformWorkspace(localWorkspace),
	)

	logger := ctx.Logger().With().Str("provider", "terraform_dir").Logger()
	runCtx := ctx.RunContext

	var policyClient *apiclient.PolicyAPIClient
	if runCtx.Config.PoliciesEnabled {
		policyClient, err = apiclient.NewPolicyAPIClient(runCtx)
		if err != nil {
			logger.Err(err).Msgf("failed to initialize policy client")
		}
	}

	var remoteCache modules.RemoteCache
	if runCtx.Config.S3ModuleCacheRegion != "" && runCtx.Config.S3ModuleCacheBucket != "" {
		s3ModuleCache, err := modules.NewS3Cache(runCtx.Config.S3ModuleCacheRegion, runCtx.Config.S3ModuleCacheBucket, runCtx.Config.S3ModuleCachePrefix, runCtx.Config.S3ModuleCachePrivate)
		if err != nil {
			logger.Warn().Msgf("failed to initialize S3 module cache: %s", err)
		} else {
			remoteCache = s3ModuleCache
		}
	}

	loader := modules.NewModuleLoader(modules.ModuleLoaderOptions{
		CachePath:           runCtx.Config.CachePath(),
		HCLParser:           modules.NewSharedHCLParser(),
		CredentialsSource:   credsSource,
		SourceMap:           runCtx.Config.TerraformSourceMap,
		SourceMapRegex:      runCtx.Config.TerraformSourceMapRegex,
		Logger:              logger,
		ModuleSync:          runCtx.ModuleMutex,
		RemoteCache:         remoteCache,
		PublicModuleChecker: modules.NewHttpPublicModuleChecker(),
	})
	cachePath := ctx.RunContext.Config.CachePath()
	initialPath := rootPath.DetectedPath
	rootPath.DetectedPath = initialPath

	if filepath.IsAbs(cachePath) {
		rootPath.DetectedPath = tryAbs(initialPath)
		rootPath.StartingPath = tryAbs(rootPath.StartingPath)
	}

	if ctx.RunContext.Config.GraphEvaluator {
		options = append(options, hcl.OptionGraphEvaluator())
	}

	if ctx.ProjectConfig.Name != "" {
		options = append(options, hcl.OptionWithProjectName(ctx.ProjectConfig.Name))
	}

	envMatcher := hcl.CreateEnvFileMatcher(ctx.RunContext.Config.Autodetect.EnvNames, ctx.RunContext.Config.Autodetect.TerraformVarFileExtensions)

	return &HCLProvider{
		policyClient:   policyClient,
		Parser:         hcl.NewParser(rootPath, envMatcher, loader, logger, options...),
		planJSONParser: NewParser(ctx, true),
		ctx:            ctx,
		config:         *config,
		logger:         logger,
	}, nil
}
func (p *HCLProvider) Context() *config.ProjectContext { return p.ctx }

func (p *HCLProvider) ProjectName() string {
	return p.Parser.ProjectName()
}

func tryAbs(initialPath string) string {
	abs, err := filepath.Abs(initialPath)
	if err != nil {
		logging.Logger.Debug().Err(err).Msgf("could not make path %s absolute", initialPath)

		return initialPath
	}

	return abs
}

func (p *HCLProvider) VarFiles() []string {
	return p.Parser.VarFiles()
}

func (p *HCLProvider) DependencyPaths() []string {
	return p.Parser.DependencyPaths()
}

func (p *HCLProvider) EnvName() string {
	return p.Parser.EnvName()
}

func (p *HCLProvider) RelativePath() string {
	return p.Parser.RelativePath()
}

func (p *HCLProvider) YAML() string {
	return p.Parser.YAML()
}

func (p *HCLProvider) Type() string        { return "terraform_dir" }
func (p *HCLProvider) DisplayType() string { return "Terraform" }
func (p *HCLProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	metadata.ConfigSha = p.ctx.ProjectConfig.ConfigSha

	modulePath := p.RelativePath()
	if modulePath != "" && modulePath != "." {
		metadata.TerraformModulePath = modulePath
	}

	metadata.TerraformWorkspace = p.ctx.ProjectConfig.TerraformWorkspace
}

// LoadResources calls a hcl.Parser to parse the directory config files into hcl.Blocks. It then builds a shallow
// representation of the terraform plan JSON files from these Blocks, this is passed to the PlanJSONProvider.
// The PlanJSONProvider uses this shallow representation to actually load Infracost resources.
func (p *HCLProvider) LoadResources(usage schema.UsageMap) ([]*schema.Project, error) {

	loadResourcesTimer := metrics.GetTimer("hcl.LoadResources", false, p.ctx.ProjectConfig.Path).Start()
	defer loadResourcesTimer.Stop()

	loadPlanTimer := metrics.GetTimer("hcl.LoadPlanJSON", false, p.ctx.ProjectConfig.Path).Start()
	j := p.LoadPlanJSON()
	loadPlanTimer.Stop()
	if j.Error != nil {
		return []*schema.Project{p.newProject(j)}, nil
	}

	project := p.newProject(j)

	parseJSONTimer := metrics.GetTimer("hcl.ParseJSON", false, p.ctx.ProjectConfig.Path).Start()
	parsedConf, err := p.planJSONParser.parseJSON(j.JSON, usage)
	parseJSONTimer.Stop()
	if err != nil {
		project.Metadata.AddError(schema.NewDiagJSONParsingFailure(err))

		return []*schema.Project{project}, nil
	}
	if p.ctx.RunContext.VCSMetadata.HasChanges() {
		j := j
		project.Metadata.VCSCodeChanged = &j.Module.HasChanges
	}

	project.AddProviderMetadata(parsedConf.ProviderMetadata)
	project.Metadata.RemoteModuleCalls = parsedConf.RemoteModuleCalls

	project.PartialPastResources = parsedConf.PastResources
	project.PartialResources = parsedConf.CurrentResources

	if p.policyClient != nil {
		uploadPolicyDataTimer := metrics.GetTimer("hcl.UploadPolicyData", false, p.ctx.ProjectConfig.Path).Start()
		err := p.policyClient.UploadPolicyData(project, parsedConf.CurrentResourceDatas, parsedConf.PastResourceDatas)
		uploadPolicyDataTimer.Stop()
		if err != nil {
			p.logger.Err(err).Msgf("failed to upload policy data %s", project.Name)
		}
	}

	return []*schema.Project{project}, nil
}

func (p *HCLProvider) newProject(parsed HCLProject) *schema.Project {
	metadata := schema.DetectProjectMetadata(parsed.Module.RootPath)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)

	if parsed.Error != nil {
		metadata.AddError(schema.NewDiagModuleEvaluationFailure(parsed.Error))
	}

	if len(parsed.Module.Warnings) > 0 {
		for _, warning := range parsed.Module.Warnings {
			p.printWarning(warning)
		}

		metadata.Warnings = append(metadata.Warnings, parsed.Module.Warnings...)
	}

	name := p.ctx.ProjectConfig.Name
	if name == "" {
		name = metadata.GenerateProjectName(p.ctx.RunContext.VCSMetadata.Remote, p.ctx.RunContext.IsCloudEnabled())

		if p.ctx.RunContext.Config.ConfigFilePath == "" && parsed.Module.ModuleSuffix != "" {
			name += "-" + parsed.Module.ModuleSuffix
		}
	}

	project := schema.NewProject(name, metadata)
	project.DisplayName = p.ProjectName()
	return project
}

func (p *HCLProvider) printWarning(warning *schema.ProjectDiag) {
	// skip warnings that don't have a friendly message
	// these are not meant to be shown to the user.
	if warning.FriendlyMessage == "" {
		return
	}

	logging.Logger.Warn().Msg(warning.FriendlyMessage)
}

type HCLProject struct {
	JSON   []byte
	Module *hcl.Module
	Error  error
}

// LoadPlanJSON parses the RootPath and return the blocks in Terraform plan JSON format.
func (p *HCLProvider) LoadPlanJSON() HCLProject {
	module := p.Module()
	if module.Error == nil {
		module.JSON, module.Error = p.modulesToPlanJSON(module.Module)
		if os.Getenv("INFRACOST_JSON_DUMP") == "true" {
			targetPath := fmt.Sprintf("%s-out.json", strings.ReplaceAll(module.Module.ModulePath, "/", "-"))
			targetDir, ok := os.LookupEnv("INFRACOST_JSON_DUMP_PATH")
			if ok {
				targetPath = filepath.Join(targetDir, targetPath)
				if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
					p.logger.Debug().Err(err).Msg("failed to create directory for json dump")
					// use the default
					targetPath = fmt.Sprintf("%s-out.json", strings.ReplaceAll(module.Module.ModulePath, "/", "-"))
				}
			}
			err := os.WriteFile(targetPath, module.JSON, os.ModePerm) // nolint: gosec
			if err != nil {
				p.logger.Debug().Err(err).Msg("failed to write to json dump")
			}
		}
	}

	return module
}

// Module parses the RootPath into an hcl Module representing a config tree of
// Terraform information. Module returns the raw hcl blocks associated with each
// found Terraform project. This can be used to fetch raw information like
// outputs, vars, resources, e.t.c.
func (p *HCLProvider) Module() HCLProject {

	metrics.GetCounter("root_module.count", true).Inc()

	if p.cache != nil {
		return *p.cache
	}

	parseTimer := metrics.GetTimer("root_module.parse.duration", false, p.Context().ProjectConfig.Path).Start()
	defer parseTimer.Stop()

	module, modErr := p.Parser.ParseDirectory()
	var v *clierror.PanicError
	if errors.As(modErr, &v) {
		err := apiclient.ReportCLIError(p.ctx.RunContext, v, false)
		if err != nil {
			p.logger.Debug().Err(err).Msg("error sending unexpected runtime error")
		}
	}

	if p.config.CacheParsingModules {
		p.cache = &HCLProject{Module: module, Error: modErr}
	}

	return HCLProject{Module: module, Error: modErr}
}

// InvalidateCache removes the module cache from the prior hcl parse.
func (p *HCLProvider) InvalidateCache() *HCLProvider {
	p.cache = nil

	return p
}

func (p *HCLProvider) newPlanSchema() {
	p.schema = &PlanSchema{
		FormatVersion:    "1.0",
		TerraformVersion: "1.1.0",
		Variables:        nil,
		PriorState: struct {
			Values PlanValues `json:"values"`
		}{
			Values: PlanValues{
				RootModule: PlanModule{
					Resources:    []ResourceJSON{},
					ChildModules: []PlanModule{},
				},
			},
		},
		InfracostResourceChanges: []ResourceChangesJSON{},
		PlannedValues: PlanValues{
			RootModule: PlanModule{
				Resources:    []ResourceJSON{},
				ChildModules: []PlanModule{},
			},
		},
		Configuration: Configuration{
			ProviderConfig: make(map[string]ProviderConfig),
			RootModule: ModuleConfig{
				Resources:   []ResourceData{},
				ModuleCalls: map[string]ModuleCall{},
			},
		},
	}
}

func (p *HCLProvider) modulesToPlanJSON(rootModule *hcl.Module) ([]byte, error) {
	p.newPlanSchema()
	if rootModule.ProviderConstraints != nil {
		p.schema.InfracostProviderConstraints = *rootModule.ProviderConstraints
	}

	mo := p.marshalModule(rootModule)
	p.schema.Configuration.RootModule = mo.ModuleConfig
	p.schema.PriorState.Values.RootModule = mo.PlanModule
	p.schema.PlannedValues.RootModule = mo.PlanModule

	b, err := json.Marshal(p.schema)
	if err != nil {
		return nil, fmt.Errorf("error handling built plan json from hcl %w", err)
	}

	return b, nil
}

func (p *HCLProvider) marshalModule(module *hcl.Module) ModuleOut {

	metrics.GetCounter("module.count", true).Inc()

	moduleConfig := ModuleConfig{
		ModuleCalls: map[string]ModuleCall{},
	}

	planModule := PlanModule{
		Address: newString(module.Name),
	}

	for _, block := range module.Blocks {
		if block.Type() == "provider" {
			p.marshalProviderBlock(block)
		}
	}

	// sort the modules so we get deterministic output in the json
	sort.SliceStable(module.Blocks, func(i, j int) bool {
		return module.Blocks[i].Label() < module.Blocks[j].Label()
	})

	configResources := map[string]struct{}{}
	for _, block := range module.Blocks {
		if block.Type() == "resource" {
			out := p.getResourceOutput(block, module.SourceURL)

			metrics.GetCounter("resource.count", true).Inc()

			if _, ok := configResources[out.Configuration.Address]; !ok {
				moduleConfig.Resources = append(moduleConfig.Resources, out.Configuration)

				configResources[out.Configuration.Address] = struct{}{}
			}

			planModule.Resources = append(planModule.Resources, out.Planned)
			p.schema.InfracostResourceChanges = append(p.schema.InfracostResourceChanges, out.Changes)
		}
	}

	// sort the modules so we get deterministic output in the json
	sort.SliceStable(module.Modules, func(i, j int) bool {
		return module.Modules[i].Name < module.Modules[j].Name
	})

	for _, m := range module.Modules {
		pieces := strings.Split(removeAddressArrayPart(m.Name), ".")
		modKey := pieces[len(pieces)-1]

		mo := p.marshalModule(m)

		moduleConfig.ModuleCalls[modKey] = ModuleCall{
			Source:       m.Source,
			ModuleConfig: mo.ModuleConfig,
			SourceUrl:    m.SourceURL,
		}

		planModule.ChildModules = append(planModule.ChildModules, mo.PlanModule)
	}

	return ModuleOut{
		PlanModule:   planModule,
		ModuleConfig: moduleConfig,
	}
}

func (p *HCLProvider) getResourceOutput(block *hcl.Block, moduleSourceURL string) ResourceOutput {
	jsonValues := marshalAttributeValues(block.Type(), block.Values())
	p.marshalBlock(block, jsonValues)
	calls := block.CallDetails()
	metadata := map[string]interface{}{
		"filename":  block.Filename,
		"startLine": block.StartLine,
		"endLine":   block.EndLine,
		"calls":     calls,
		"checksum":  generateChecksum(jsonValues),
	}
	if attrsWithMissingKeys := block.AttributesWithUnknownKeys(); len(attrsWithMissingKeys) > 0 {
		metadata["attributesWithUnknownKeys"] = attrsWithMissingKeys
	}

	// if there is a non-empty module source url, this means that resource comes from a remote module
	// in this case we should add the module source url to the metadata.
	if moduleSourceURL != "" && moduleCacheRegex.MatchString(block.Filename) {
		filename := buildModuleFilename(block.Filename, moduleSourceURL)
		if filename != "" {
			metadata["moduleFilename"] = filename
		}
	}

	planned := ResourceJSON{
		Address:           block.FullName(),
		Mode:              "managed",
		Type:              block.TypeLabel(),
		Name:              stripCountOrForEach(block.NameLabel()),
		Index:             block.Index(),
		SchemaVersion:     0,
		InfracostMetadata: metadata,
	}

	changes := ResourceChangesJSON{
		Address:       block.FullName(),
		ModuleAddress: newString(block.ModuleAddress()),
		Mode:          "managed",
		Type:          block.TypeLabel(),
		Name:          stripCountOrForEach(block.NameLabel()),
		Index:         block.Index(),
		Change: ResourceChange{
			Actions: []string{"create"},
		},
	}

	planned.Values = jsonValues
	changes.Change.After = jsonValues

	configuration := ResourceData{
		Address:           stripCountOrForEach(block.LocalName()),
		Mode:              "managed",
		Type:              block.TypeLabel(),
		Name:              stripCountOrForEach(block.NameLabel()),
		ProviderConfigKey: block.ProviderConfigKey(),
		Expressions:       blockToReferences(block),
		CountExpression:   p.countReferences(block),
	}

	return ResourceOutput{
		Planned:       planned,
		PriorState:    planned,
		Changes:       changes,
		Configuration: configuration,
	}
}

func buildModuleFilename(filename string, moduleSourceURL string) string {
	httpsURL, err := normalizeModuleURL(moduleSourceURL)
	if err != nil {
		logging.Logger.Debug().Err(err).Msgf("failed to build module filename, could not transform url %s to https", moduleSourceURL)
		return ""
	}

	u, err := url.Parse(httpsURL)
	if err != nil {
		logging.Logger.Debug().Err(err).Msgf("failed to build module filename,could not parse url %s", httpsURL)
		return ""
	}

	ref := "HEAD"
	queryRef := u.Query().Get("ref")
	if queryRef != "" {
		ref = queryRef
	}
	u.Path += "/blob/" + ref + "/"
	u.RawQuery = ""

	matches := moduleCacheRegex.FindStringSubmatch(filename)
	moduleFilename := u.String() + matches[1]
	return moduleFilename
}

func normalizeModuleURL(sshURL string) (string, error) {
	// git::ssh and git:https aren't recognized as a valid URL scheme, so we need to strip it so they just use ssh or https schemes
	u := sshURL
	if strings.HasPrefix(sshURL, "git::") {
		u = strings.Replace(sshURL, "git::", "", 1)
	}

	parsedURL, err := url.Parse(u)
	if err != nil {
		// Try parsing it as a git URL
		parsedURL, err = giturls.Parse(u)
		if err != nil {
			return "", err
		}
	}

	// Save the query string, since this is lost when we normalize the URL
	// and it may contain the ref
	query := parsedURL.Query()

	res, err := modules.NormalizeGitURLToHTTPS(parsedURL)
	if err != nil {
		return "", err
	}

	// Add the query string back in
	res.RawQuery = query.Encode()

	return res.String(), nil
}

func (p *HCLProvider) marshalProviderBlock(block *hcl.Block) string {
	providerConfigKey := block.Values().GetAttr("config_key").AsString()

	providerType := block.TypeLabel()
	name := providerType
	if a := block.GetAttribute("alias"); a != nil {
		name = name + "." + a.AsString()
	}

	region := block.GetAttribute("region").AsString()

	metadata := map[string]interface{}{
		"filename":   block.Filename,
		"start_line": block.StartLine,
		"end_line":   block.EndLine,
	}

	if attrsWithMissingKeys := block.AttributesWithUnknownKeys(); len(attrsWithMissingKeys) > 0 {
		metadata["attributes_with_unknown_keys"] = attrsWithMissingKeys
	}

	p.schema.Configuration.ProviderConfig[providerConfigKey] = ProviderConfig{
		Name: name,
		Expressions: map[string]interface{}{
			"region": map[string]interface{}{
				"constant_value": region,
			},
		},
		InfracostMetadata: metadata,
	}

	switch providerType {
	case "aws":
		if defaultTags := p.marshalAWSDefaultTagsBlock(block); defaultTags != nil {
			p.schema.Configuration.ProviderConfig[providerConfigKey].Expressions["default_tags"] = []map[string]interface{}{defaultTags}
		}
	case "google":
		if defaultLabels := p.marshalGoogleDefaultTagsBlock(block); defaultLabels != nil {
			p.schema.Configuration.ProviderConfig[providerConfigKey].Expressions["default_labels"] = defaultLabels
		}
	}

	return name
}

func (p *HCLProvider) marshalAWSDefaultTagsBlock(providerBlock *hcl.Block) map[string]interface{} {
	b := providerBlock.GetChildBlock("default_tags")
	if b == nil {
		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			p.logger.Debug().Msgf("could not marshal default_tags block: %v", r)
		}
	}()

	marshalledTags := make(map[string]interface{})

	tags := b.GetAttribute("tags")
	if tags == nil {
		return marshalledTags
	}

	value := tags.Value()
	if value.IsNull() || !value.IsKnown() || !value.CanIterateElements() {
		return nil
	}

	for tag, val := range value.AsValueMap() {
		if !val.IsKnown() {
			p.logger.Debug().Msgf("tag %s has unknown value, cannot marshal", tag)
			continue
		}

		if val.Type().Equals(cty.Bool) {
			var tagValue bool
			err := gocty.FromCtyValue(val, &tagValue)
			if err != nil {
				p.logger.Debug().Err(err).Msgf("could not marshal tag %s to bool value", tag)
				continue
			}

			marshalledTags[tag] = fmt.Sprintf("%t", tagValue)
			continue
		}

		if val.Type() == cty.Number {
			var tagValue big.Float
			err := gocty.FromCtyValue(val, &tagValue)
			if err != nil {
				p.logger.Debug().Err(err).Msgf("could not marshal tag %s to number value", tag)
				continue
			}

			marshalledTags[tag] = tagValue.String()
			continue
		}

		var tagValue string
		err := gocty.FromCtyValue(val, &tagValue)
		if err != nil {
			p.logger.Debug().Err(err).Msgf("could not marshal tag %s to string value", tag)
			continue
		}

		marshalledTags[tag] = tagValue
	}

	tagsVal := map[string]interface{}{
		"constant_value": marshalledTags,
	}

	if refs := tags.ReferencesCausingUnknownKeys(); len(refs) > 0 {
		tagsVal["missing_attributes_causing_unknown_keys"] = refs
	}

	return map[string]interface{}{
		"tags": tagsVal,
	}
}

func (p *HCLProvider) marshalGoogleDefaultTagsBlock(providerBlock *hcl.Block) map[string]interface{} {
	tags := providerBlock.GetAttribute("default_labels")
	if tags == nil {
		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			p.logger.Debug().Msgf("could not marshal default_labels block: %v", r)
		}
	}()

	marshalledTags := make(map[string]interface{})

	value := tags.Value()
	if value.IsNull() || !value.IsKnown() || !value.CanIterateElements() {
		return nil
	}

	for tag, val := range value.AsValueMap() {
		if !val.IsKnown() {
			p.logger.Debug().Msgf("tag %s has unknown value, cannot marshal", tag)
			continue
		}

		if val.Type().Equals(cty.Bool) {
			var tagValue bool
			err := gocty.FromCtyValue(val, &tagValue)
			if err != nil {
				p.logger.Debug().Err(err).Msgf("could not marshal tag %s to bool value", tag)
				continue
			}

			marshalledTags[tag] = fmt.Sprintf("%t", tagValue)
			continue
		}

		if val.Type() == cty.Number {
			var tagValue big.Float
			err := gocty.FromCtyValue(val, &tagValue)
			if err != nil {
				p.logger.Debug().Err(err).Msgf("could not marshal tag %s to number value", tag)
				continue
			}

			marshalledTags[tag] = tagValue.String()
			continue
		}

		var tagValue string
		err := gocty.FromCtyValue(val, &tagValue)
		if err != nil {
			p.logger.Debug().Err(err).Msgf("could not marshal tag %s to string value", tag)
			continue
		}

		marshalledTags[tag] = tagValue
	}

	tagsVal := map[string]interface{}{
		"constant_value": marshalledTags,
	}
	if refs := tags.ReferencesCausingUnknownKeys(); len(refs) > 0 {
		tagsVal["missing_attributes_causing_unknown_keys"] = refs
	}
	return tagsVal
}

func (p *HCLProvider) countReferences(block *hcl.Block) *countExpression {
	for _, attribute := range block.GetAttributes() {
		name := attribute.Name()
		if name != "count" {
			continue
		}

		exp := countExpression{}

		references := attribute.AllReferences()
		if len(references) > 0 {
			for _, ref := range references {
				exp.References = append(
					exp.References,
					strings.ReplaceAll(ref.String(), "variable", "var"),
				)
			}

			return &exp
		}

		i := attribute.AsInt()
		exp.ConstantValue = &i
		return &exp
	}

	return nil
}

var ignoredAttrs = map[string]bool{"arn": true, "id": true, "name": true, "self_link": true}
var checksumMarshaller = jsoniter.ConfigCompatibleWithStandardLibrary

func generateChecksum(value map[string]interface{}) string {
	filtered := make(map[string]interface{})
	for k, v := range value {
		if !ignoredAttrs[k] {
			filtered[k] = v
		}
	}

	serialized, err := checksumMarshaller.Marshal(filtered)
	if err != nil {
		return ""
	}

	h := sha256.New()
	h.Write(serialized)

	return hex.EncodeToString(h.Sum(nil))
}

func blockToReferences(block *hcl.Block) map[string]interface{} {
	expressionValues := make(map[string]interface{})

	for _, attribute := range block.GetAttributes() {
		references := attribute.AllReferences()
		if len(references) > 0 {
			r := refs{}
			for _, ref := range references {
				r.References = append(r.References, ref.JSONString())
			}

			// counts are special expressions that have their own json key.
			// So we ignore them here.
			name := attribute.Name()
			if name == "count" {
				continue
			}

			expressionValues[name] = r
		}

		childExpressions := make(map[string][]interface{})
		for _, child := range block.Children() {
			vals := childExpressions[child.Type()]
			childReferences := blockToReferences(child)

			if len(childReferences) > 0 {
				childExpressions[child.Type()] = append(vals, childReferences)
			}
		}

		if len(childExpressions) > 0 {
			for name, v := range childExpressions {
				expressionValues[name] = v
			}
		}
	}

	return expressionValues
}

func (p *HCLProvider) marshalBlock(block *hcl.Block, jsonValues map[string]interface{}) {
	for _, b := range block.Children() {
		key := b.Type()
		if key == "dynamic" {
			continue
		}

		childValues := marshalAttributeValues(key, b.Values())
		if len(b.Children()) > 0 {
			p.marshalBlock(b, childValues)
		}

		if v, ok := jsonValues[key]; ok {
			if _, ok := v.(stdJson.RawMessage); ok {
				p.logger.Debug().
					Str("parent_block", block.LocalName()).
					Str("child_block", b.LocalName()).
					Msgf("skipping attribute '%s' that has also been declared as a child block", key)

				continue
			}

			jsonValues[key] = append(v.([]interface{}), childValues)
			continue
		}

		jsonValues[key] = []interface{}{childValues}
	}
}

func marshalAttributeValues(blockType string, value cty.Value) map[string]interface{} {
	if value.IsNull() {
		return nil
	}
	ret := make(map[string]interface{})

	it := value.ElementIterator()
	for it.Next() {
		k, v := it.Element()
		vJSON, _ := ctyJson.Marshal(v, v.Type())
		var key string
		err := gocty.FromCtyValue(k, &key)
		if err != nil {
			logging.Logger.Debug().Err(err).Msgf("could not convert block map key to string ignoring entry")
			continue
		}

		if (blockType == "resource" || blockType == "module") && key == "count" {
			continue
		}

		ret[key] = stdJson.RawMessage(vJSON)
	}
	return ret
}

type ResourceOutput struct {
	Planned       ResourceJSON
	Changes       ResourceChangesJSON
	PriorState    ResourceJSON
	Configuration ResourceData
}

type ResourceJSON struct {
	Address           string                 `json:"address"`
	Mode              string                 `json:"mode"`
	Type              string                 `json:"type"`
	Name              string                 `json:"name"`
	Index             *int64                 `json:"index,omitempty"`
	SchemaVersion     int                    `json:"schema_version"`
	Values            map[string]interface{} `json:"values"`
	InfracostMetadata map[string]interface{} `json:"infracost_metadata"`
}

type ResourceChangesJSON struct {
	Address       string         `json:"address"`
	ModuleAddress *string        `json:"module_address,omitempty"`
	Mode          string         `json:"mode"`
	Type          string         `json:"type"`
	Name          string         `json:"name"`
	Index         *int64         `json:"index,omitempty"`
	Change        ResourceChange `json:"change"`
}

type ResourceChange struct {
	Actions []string               `json:"actions"`
	Before  interface{}            `json:"before"`
	After   map[string]interface{} `json:"after"`
}

type PlanValues struct {
	RootModule PlanModule `json:"root_module"`
}

type PlanSchema struct {
	FormatVersion    string      `json:"format_version"`
	TerraformVersion string      `json:"terraform_version"`
	Variables        interface{} `json:"variables,omitempty"`
	PriorState       struct {
		Values PlanValues `json:"values"`
	} `json:"prior_state"`
	PlannedValues PlanValues    `json:"planned_values"`
	Configuration Configuration `json:"configuration"`

	// InfracostResourceChanges is a flattened list of resource changes for the plan, this is in the format of the Terraform
	// plan JSON output, but we omit adding it as the supported `resource_changes` key as this will cause plan inconsistencies.
	// We copy this `infracost_resource_changes` key at a later date to `resource_changes` before sending to the Policy API.
	// This means that we can evaluate the Rego ruleset on the known Terraform plan JSON structure.
	InfracostResourceChanges     []ResourceChangesJSON   `json:"infracost_resource_changes"`
	InfracostProviderConstraints hcl.ProviderConstraints `json:"infracost_provider_constraints"`
}

type PlanModule struct {
	Resources    []ResourceJSON `json:"resources,omitempty"`
	Address      *string        `json:"address,omitempty"`
	ChildModules []PlanModule   `json:"child_modules,omitempty"`
}

type Configuration struct {
	ProviderConfig map[string]ProviderConfig `json:"provider_config"`
	RootModule     ModuleConfig              `json:"root_module"`
}

type ModuleConfig struct {
	Resources   []ResourceData        `json:"resources,omitempty"`
	ModuleCalls map[string]ModuleCall `json:"module_calls,omitempty"`
}

type ModuleOut struct {
	PlanModule   PlanModule
	ModuleConfig ModuleConfig
}

type ProviderConfig struct {
	Name              string                 `json:"name"`
	Expressions       map[string]interface{} `json:"expressions,omitempty"`
	InfracostMetadata map[string]interface{} `json:"infracost_metadata"`
}

type ResourceData struct {
	Address           string                 `json:"address"`
	Mode              string                 `json:"mode"`
	Type              string                 `json:"type"`
	Name              string                 `json:"name"`
	ProviderConfigKey string                 `json:"provider_config_key"`
	Expressions       map[string]interface{} `json:"expressions,omitempty"`
	SchemaVersion     int                    `json:"schema_version"`
	CountExpression   *countExpression       `json:"count_expression,omitempty"`
}

type ModuleCall struct {
	Source       string       `json:"source"`
	ModuleConfig ModuleConfig `json:"module"`
	SourceUrl    string       `json:"sourceUrl,omitempty"`
}

type countExpression struct {
	References    []string `json:"references,omitempty"`
	ConstantValue *int64   `json:"constant_value,omitempty"`
}

type refs struct {
	References []string `json:"references"`
}

func newString(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}

var countRegex = regexp.MustCompile(`\[.+\]$`)

func stripCountOrForEach(s string) string {
	return countRegex.ReplaceAllString(s, "")
}
