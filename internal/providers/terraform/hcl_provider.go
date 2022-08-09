package terraform

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	ctyJson "github.com/zclconf/go-cty/cty/json"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/hcl"
	"github.com/infracost/infracost/internal/hcl/modules"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
)

type HCLProvider struct {
	parsers        []*hcl.Parser
	planJSONParser *Parser
	logger         *log.Entry

	schema      *PlanSchema
	providerKey string
	ctx         *config.ProjectContext
	cache       []*hcl.Module
	config      HCLProviderConfig
}

type HCLProviderConfig struct {
	SuppressLogging     bool
	CacheParsingModules bool
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
func NewHCLProvider(ctx *config.ProjectContext, config *HCLProviderConfig, opts ...hcl.Option) (*HCLProvider, error) {
	if config == nil {
		config = &HCLProviderConfig{}
	}

	v, err := varsFromPlanFlags(ctx.ProjectConfig.TerraformPlanFlags)
	if err != nil {
		return nil, fmt.Errorf("could not parse vars from plan flags %w", err)
	}

	options := []hcl.Option{hcl.OptionWithTFEnvVars(ctx.ProjectConfig.Env)}

	if len(v.vars) > 0 {
		withPlanFlagVars := hcl.OptionWithPlanFlagVars(v.vars)
		options = append(options, withPlanFlagVars)
	}

	v.files = append(v.files, ctx.ProjectConfig.TerraformVarFiles...)
	if len(v.files) > 0 {
		withFiles := hcl.OptionWithTFVarsPaths(v.files)
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
		options = append(options, hcl.OptionWithRemoteVarLoader(
			credsSource.BaseCredentialSet.Host,
			credsSource.BaseCredentialSet.Token,
			localWorkspace),
		)
	}

	options = append(options,
		hcl.OptionWithTerraformWorkspace(localWorkspace),
		hcl.OptionWithCredentialsSource(credsSource),
	)

	logger := ctx.Logger().WithFields(log.Fields{"provider": "terraform_dir"})
	parsers, err := hcl.LoadParsers(ctx.ProjectConfig.Path, ctx.ProjectConfig.ExcludePaths, logger, options...)
	if err != nil {
		return nil, err
	}

	return &HCLProvider{
		parsers:        parsers,
		planJSONParser: NewParser(ctx, false),
		ctx:            ctx,
		config:         *config,
		logger:         logger,
	}, err
}

func (p *HCLProvider) Type() string        { return "terraform_dir" }
func (p *HCLProvider) DisplayType() string { return "Terraform directory" }
func (p *HCLProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	basePath := p.ctx.ProjectConfig.Path
	if p.ctx.RunContext.Config.ConfigFilePath != "" {
		basePath = filepath.Dir(p.ctx.RunContext.Config.ConfigFilePath)
	}

	modulePath, err := filepath.Rel(basePath, metadata.Path)
	if err == nil && modulePath != "" && modulePath != "." {
		p.logger.Debugf("calculated relative terraformModulePath for %s from %s", basePath, metadata.Path)
		metadata.TerraformModulePath = modulePath
	}

	metadata.TerraformWorkspace = p.ctx.ProjectConfig.TerraformWorkspace
}

// LoadResources calls a hcl.Parser to parse the directory config files into hcl.Blocks. It then builds a shallow
// representation of the terraform plan JSON files from these Blocks, this is passed to the PlanJSONProvider.
// The PlanJSONProvider uses this shallow representation to actually load Infracost resources.
func (p *HCLProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	jsons, err := p.LoadPlanJSONs()
	if err != nil {
		return nil, err
	}

	var projects = make([]*schema.Project, len(jsons))
	for i, j := range jsons {
		project, err := p.parseResources(j.path, j.json, usage)
		if err != nil {
			return nil, err
		}

		projects[i] = project
	}

	return projects, nil
}

func (p *HCLProvider) parseResources(path string, j []byte, usage map[string]*schema.UsageData) (*schema.Project, error) {
	metadata := config.DetectProjectMetadata(path)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)
	name := schema.GenerateProjectName(metadata, p.ctx.ProjectConfig.Name, p.ctx.RunContext.IsCloudEnabled())

	project := schema.NewProject(name, metadata)

	pastResources, resources, err := p.planJSONParser.parseJSON(j, usage)
	if err != nil {
		return project, fmt.Errorf("Error parsing Terraform plan JSON file %w", err)
	}

	project.PastResources = pastResources
	project.Resources = resources

	return project, nil
}

type hclProject struct {
	path string
	json []byte
}

// LoadPlanJSONs parses the found directories and return the blocks in Terraform plan JSON format.
func (p *HCLProvider) LoadPlanJSONs() ([]hclProject, error) {
	var jsons = make([]hclProject, len(p.parsers))
	modules, err := p.Modules()
	if err != nil {
		return nil, err
	}

	for i, module := range modules {
		b, err := p.modulesToPlanJSON(module)
		if err != nil {
			return nil, err
		}

		jsons[i] = hclProject{json: b, path: module.RootPath}
	}

	return jsons, nil
}

// Modules parses the found directories into hcl modules representing a config tree of Terraform information.
// Modules returns the raw hcl blocks associated with each found Terraform project. This can be used
// to fetch raw information like outputs, vars, resources, e.t.c.
func (p *HCLProvider) Modules() ([]*hcl.Module, error) {
	if p.cache != nil {
		return p.cache, nil
	}

	var modules = make([]*hcl.Module, len(p.parsers))

	for i, parser := range p.parsers {
		if len(p.parsers) > 1 && !p.config.SuppressLogging {
			fmt.Fprintf(os.Stderr, "Detected Terraform project at %s\n", ui.DisplayPath(parser.Path()))
		}

		module, err := parser.ParseDirectory()
		if err != nil {
			return nil, err
		}

		modules[i] = module
	}

	if p.config.CacheParsingModules {
		p.cache = modules
	}

	return modules, nil
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
		PlannedValues: struct {
			RootModule PlanModule `json:"root_module"`
		}{
			RootModule: PlanModule{
				Resources:    []ResourceJSON{},
				ChildModules: []PlanModule{},
			},
		},
		ResourceChanges: []ResourceChangesJSON{},
		Configuration: Configuration{
			ProviderConfig: make(map[string]ProviderConfig),
			RootModule: ModuleConfig{
				Resources:   []ResourceData{},
				ModuleCalls: map[string]ModuleCall{},
			},
		},
	}

	p.providerKey = ""
}

func (p *HCLProvider) modulesToPlanJSON(rootModule *hcl.Module) ([]byte, error) {
	p.newPlanSchema()

	mo := p.marshalModule(rootModule)
	p.schema.Configuration.RootModule = mo.ModuleConfig
	p.schema.PlannedValues.RootModule = mo.PlanModule

	b, err := json.MarshalIndent(p.schema, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error handling built plan json from hcl %w", err)
	}
	return b, nil
}

func (p *HCLProvider) marshalModule(module *hcl.Module) ModuleOut {
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

	configResources := map[string]struct{}{}
	for _, block := range module.Blocks {
		if block.Type() == "resource" {
			out := p.getResourceOutput(block)

			if _, ok := configResources[out.Configuration.Address]; !ok {
				moduleConfig.Resources = append(moduleConfig.Resources, out.Configuration)

				configResources[out.Configuration.Address] = struct{}{}
			}

			planModule.Resources = append(planModule.Resources, out.Planned)

			p.schema.ResourceChanges = append(p.schema.ResourceChanges, out.Changes)
		}
	}

	for _, m := range module.Modules {
		pieces := strings.Split(m.Name, ".")
		modKey := pieces[len(pieces)-1]

		mo := p.marshalModule(m)

		moduleConfig.ModuleCalls[modKey] = ModuleCall{
			Source:       m.Source,
			ModuleConfig: mo.ModuleConfig,
		}

		planModule.ChildModules = append(planModule.ChildModules, mo.PlanModule)
	}

	return ModuleOut{
		PlanModule:   planModule,
		ModuleConfig: moduleConfig,
	}
}

func (p *HCLProvider) getResourceOutput(block *hcl.Block) ResourceOutput {
	planned := ResourceJSON{
		Address:       block.FullName(),
		Mode:          "managed",
		Type:          block.TypeLabel(),
		Name:          stripCount(block.NameLabel()),
		Index:         block.Index(),
		SchemaVersion: 0,
		InfracostMetadata: map[string]interface{}{
			"filename": block.Filename,
			"calls":    block.CallDetails(),
		},
	}

	changes := ResourceChangesJSON{
		Address:       block.FullName(),
		ModuleAddress: newString(block.ModuleAddress()),
		Mode:          "managed",
		Type:          block.TypeLabel(),
		Name:          stripCount(block.NameLabel()),
		Index:         block.Index(),
		Change: ResourceChange{
			Actions: []string{"create"},
		},
	}

	jsonValues := marshalAttributeValues(block.Type(), block.Values())
	p.marshalBlock(block, jsonValues)

	changes.Change.After = jsonValues
	planned.Values = jsonValues

	providerConfigKey := p.providerKey
	providerAttr := block.GetAttribute("provider")
	if providerAttr != nil {
		r, err := providerAttr.Reference()
		if err == nil {
			providerConfigKey = r.String()
		}

		value := providerAttr.AsString()
		if err != nil && value != "" {
			providerConfigKey = value
		}
	}

	var configuration ResourceData
	if block.HasModuleBlock() {
		configuration = ResourceData{
			Address:           stripCount(block.LocalName()),
			Mode:              "managed",
			Type:              block.TypeLabel(),
			Name:              stripCount(block.NameLabel()),
			ProviderConfigKey: block.ModuleName() + ":" + block.Provider(),
			Expressions:       blockToReferences(block),
			CountExpression:   p.countReferences(block),
		}
	} else {
		configuration = ResourceData{
			Address:           stripCount(block.FullName()),
			Mode:              "managed",
			Type:              block.TypeLabel(),
			Name:              stripCount(block.NameLabel()),
			ProviderConfigKey: providerConfigKey,
			Expressions:       blockToReferences(block),
			CountExpression:   p.countReferences(block),
		}
	}

	return ResourceOutput{
		Planned:       planned,
		Changes:       changes,
		Configuration: configuration,
	}
}

func (p *HCLProvider) marshalProviderBlock(block *hcl.Block) string {
	name := block.TypeLabel()
	if a := block.GetAttribute("alias"); a != nil {
		name = name + "." + a.AsString()
	}

	region := block.GetAttribute("region").AsString()

	p.schema.Configuration.ProviderConfig[name] = ProviderConfig{
		Name: name,
		Expressions: map[string]interface{}{
			"region": map[string]interface{}{
				"constant_value": region,
			},
		},
	}

	if p.providerKey == "" {
		p.providerKey = name
	}

	return name
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
		if key == "dynamic" || key == "depends_on" {
			continue
		}

		childValues := marshalAttributeValues(key, b.Values())
		if len(b.Children()) > 0 {
			p.marshalBlock(b, childValues)
		}

		if v, ok := jsonValues[key]; ok {
			if _, ok := v.(json.RawMessage); ok {
				p.logger.WithFields(log.Fields{
					"parent_block": block.LocalName(),
					"child_block":  b.LocalName(),
				}).Debugf("skipping attribute '%s' that has also been declared as a child block", key)

				continue
			}

			jsonValues[key] = append(v.([]interface{}), childValues)
			continue
		}

		jsonValues[key] = []interface{}{childValues}
	}
}

func marshalAttributeValues(blockType string, value cty.Value) map[string]interface{} {
	if value == cty.NilVal || value.IsNull() {
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
			logging.Logger.WithError(err).Debugf("could not convert block map key to string ignoring entry")
			continue
		}

		if (blockType == "resource" || blockType == "module") && key == "count" {
			continue
		}

		ret[key] = json.RawMessage(vJSON)
	}
	return ret
}

type ResourceOutput struct {
	Planned       ResourceJSON
	Changes       ResourceChangesJSON
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

type PlanSchema struct {
	FormatVersion    string      `json:"format_version"`
	TerraformVersion string      `json:"terraform_version"`
	Variables        interface{} `json:"variables,omitempty"`
	PlannedValues    struct {
		RootModule PlanModule `json:"root_module"`
	} `json:"planned_values"`
	ResourceChanges []ResourceChangesJSON `json:"resource_changes"`
	Configuration   Configuration         `json:"configuration"`
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
	Name        string                 `json:"name"`
	Expressions map[string]interface{} `json:"expressions,omitempty"`
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

var countRegex = regexp.MustCompile(`\[\d+\]$`)

func stripCount(s string) string {
	return countRegex.ReplaceAllString(s, "")
}
