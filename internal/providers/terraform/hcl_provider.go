package terraform

import (
	"encoding/json"
	"fmt"

	"github.com/zclconf/go-cty/cty"
	ctyJson "github.com/zclconf/go-cty/cty/json"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/hcl"
	"github.com/infracost/infracost/internal/schema"
)

type HCLProvider struct {
	Parser   *hcl.Parser
	Provider *PlanJSONProvider
}

func NewHCLProvider(ctx *config.ProjectContext, provider *PlanJSONProvider) (HCLProvider, error) {
	option, err := hcl.TfVarsToOption(ctx.ProjectConfig.TFVarFiles...)
	if err != nil {
		return HCLProvider{}, err
	}

	p := hcl.New(ctx.ProjectConfig.Path, option)

	return HCLProvider{
		Parser:   p,
		Provider: provider,
	}, err
}

func (p HCLProvider) Type() string {
	return "hcl_provider"
}

func (p HCLProvider) DisplayType() string {
	return "HCL Provider"
}

func (p HCLProvider) AddMetadata(metadata *schema.ProjectMetadata) {
}

func (p HCLProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	b, err := p.LoadPlanJSON()
	if err != nil {
		return nil, err
	}

	return p.Provider.LoadResourcesFromSrc(usage, b, nil)
}

func (p HCLProvider) LoadPlanJSON() ([]byte, error) {
	modules, err := p.Parser.ParseDirectory()
	if err != nil {
		return nil, err
	}

	sch := p.modulesToPlanJSON(modules)
	b, err := json.Marshal(sch)
	if err != nil {
		return nil, fmt.Errorf("error handling built plan json from hcl %w", err)
	}
	return b, nil
}

func (p HCLProvider) modulesToPlanJSON(modules []*hcl.Module) PlanSchema {
	sch := PlanSchema{
		FormatVersion:    "1.0",
		TerraformVersion: "1.1.0",
		Variables:        nil,
		PlannedValues: struct {
			RootModule PlanRootModule `json:"root_module"`
		}{
			RootModule: PlanRootModule{
				Resources:    []ResourceJSON{},
				ChildModules: []ChildModule{{}},
			},
		},
		ResourceChanges: []ResourceChangesJSON{},
		Configuration: Configuration{
			ProviderConfig: make(map[string]ProviderConfig),
			RootModule: struct {
				Resources   []ResourceData        `json:"resources,omitempty"`
				ModuleCalls map[string]ModuleCall `json:"module_calls"`
			}{
				Resources:   []ResourceData{},
				ModuleCalls: map[string]ModuleCall{},
			},
		},
	}

	for _, module := range modules {
		var providerKey string

		for _, block := range module.Blocks {
			if block.Type() == "provider" {
				name := block.TypeLabel()
				if a := block.GetAttribute("alias"); a != nil {
					name = a.Value().AsString()
				}

				// set the default provider key
				if providerKey == "" {
					providerKey = name
				}

				region := ""
				value := block.GetAttribute("region").Value()
				if value != cty.NilVal {
					region = value.AsString()
				}

				sch.Configuration.ProviderConfig[name] = ProviderConfig{
					Name: name,
					Expressions: map[string]interface{}{
						"region": map[string]interface{}{
							"constant_value": region,
						},
					},
				}
			}
		}

		for _, block := range module.Blocks {
			if block.Type() == "resource" {
				r := ResourceJSON{
					Address:       block.FullName(),
					Mode:          "managed",
					Type:          block.TypeLabel(),
					Name:          block.NameLabel(),
					ProviderName:  "registry.terraform.io/hashicorp/aws",
					SchemaVersion: 1,
				}

				c := ResourceChangesJSON{
					Address:       block.FullName(),
					ModuleAddress: block.ModuleAddress(),
					Mode:          "managed",
					Type:          block.TypeLabel(),
					Name:          block.NameLabel(),
					ProviderName:  "registry.terraform.io/hashicorp/aws",
					Change: struct {
						Actions []string               `json:"actions"`
						Before  interface{}            `json:"before"`
						After   map[string]interface{} `json:"after"`
					}{
						Actions: []string{"create"},
					},
				}

				jsonValues := marshalAttributeValues(block.Type(), block.Values())
				marshalBlock(block, jsonValues)

				c.Change.After = jsonValues
				r.Values = jsonValues

				providerConfigKey := providerKey
				providerAttr := block.GetAttribute("provider")
				if providerAttr != nil {
					value := providerAttr.Value()
					if value.Type() == cty.String {
						providerConfigKey = value.AsString()
					}
				}

				if block.HasModuleBlock() {
					modCall, ok := sch.Configuration.RootModule.ModuleCalls[block.ModuleName()]
					if !ok {
						modCall = ModuleCall{
							Source: block.ModuleSource(),
							Module: ModuleCallModule{
								Resources: []ResourceData{},
							},
						}
					}

					modCall.Module.Resources = append(modCall.Module.Resources, ResourceData{
						Address:           block.LocalName(),
						Mode:              "managed",
						Type:              block.TypeLabel(),
						Name:              block.NameLabel(),
						ProviderConfigKey: block.ModuleName() + ":" + block.Provider(),
						Expressions:       blockToReferences(block), // This doesn't seem to work for module calls, but it is not clear that it is needed.
					})
					sch.Configuration.RootModule.ModuleCalls[block.ModuleName()] = modCall

					sch.PlannedValues.RootModule.ChildModules[0].Resources = append(sch.PlannedValues.RootModule.ChildModules[0].Resources, r)
				} else {
					sch.Configuration.RootModule.Resources = append(sch.Configuration.RootModule.Resources, ResourceData{
						Address:           block.FullName(),
						Mode:              "managed",
						Type:              block.TypeLabel(),
						Name:              block.LocalName(),
						ProviderConfigKey: providerConfigKey,
						Expressions:       blockToReferences(block),
					})

					sch.PlannedValues.RootModule.Resources = append(sch.PlannedValues.RootModule.Resources, r)
				}

				sch.ResourceChanges = append(sch.ResourceChanges, c)
			}
		}
	}

	return sch
}

func blockToReferences(block *hcl.Block) map[string]interface{} {
	expressionValues := make(map[string]interface{})

	for _, attribute := range block.GetAttributes() {
		references := attribute.AllReferences()
		if len(references) > 0 {
			r := refs{}
			for _, ref := range references {
				r.References = append(r.References, ref.String())
			}

			expressionValues[attribute.Name()] = r
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

func marshalBlock(block *hcl.Block, jsonValues map[string]interface{}) {
	for _, b := range block.Children() {
		childValues := marshalAttributeValues(b.Type(), b.Values())
		if len(b.Children()) > 0 {
			marshalBlock(b, childValues)
		}

		if v, ok := jsonValues[b.Type()]; ok {
			jsonValues[b.Type()] = append(v.([]interface{}), childValues)
			continue
		}

		jsonValues[b.Type()] = []interface{}{childValues}
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
		key := k.AsString()

		if (blockType == "resource" || blockType == "module") && key == "count" {
			continue
		}

		ret[key] = json.RawMessage(vJSON)
	}
	return ret
}

type ResourceJSON struct {
	Address       string                 `json:"address"`
	Mode          string                 `json:"mode"`
	Type          string                 `json:"type"`
	Name          string                 `json:"name"`
	ProviderName  string                 `json:"provider_name"`
	SchemaVersion int                    `json:"schema_version"`
	Values        map[string]interface{} `json:"values"`
}

type ResourceChangesJSON struct {
	Address       string `json:"address"`
	ModuleAddress string `json:"module_address"`
	Mode          string `json:"mode"`
	Type          string `json:"type"`
	Name          string `json:"name"`
	ProviderName  string `json:"provider_name"`
	Change        struct {
		Actions []string               `json:"actions"`
		Before  interface{}            `json:"before"`
		After   map[string]interface{} `json:"after"`
	} `json:"change"`
}

type PlanSchema struct {
	FormatVersion    string      `json:"format_version"`
	TerraformVersion string      `json:"terraform_version"`
	Variables        interface{} `json:"variables,omitempty"`
	PlannedValues    struct {
		RootModule PlanRootModule `json:"root_module"`
	} `json:"planned_values"`
	ResourceChanges []ResourceChangesJSON `json:"resource_changes"`
	Configuration   Configuration         `json:"configuration"`
}

type PlanRootModule struct {
	Resources    []ResourceJSON `json:"resources,omitempty"`
	ChildModules []ChildModule  `json:"child_modules"`
}

type Configuration struct {
	ProviderConfig map[string]ProviderConfig `json:"provider_config"`
	RootModule     struct {
		Resources   []ResourceData        `json:"resources,omitempty"`
		ModuleCalls map[string]ModuleCall `json:"module_calls"`
	} `json:"root_module"`
}

type ProviderConfig struct {
	Name        string                 `json:"name"`
	Expressions map[string]interface{} `json:"expressions"`
}

type ResourceData struct {
	Address           string                 `json:"address"`
	Mode              string                 `json:"mode"`
	Type              string                 `json:"type"`
	Name              string                 `json:"name"`
	ProviderConfigKey string                 `json:"provider_config_key"`
	Expressions       map[string]interface{} `json:"expressions"`
}

type ModuleCall struct {
	Source string           `json:"source"`
	Module ModuleCallModule `json:"module"`
}

type ModuleCallModule struct {
	Resources []ResourceData `json:"resources"`
}

type ChildModule struct {
	Resources []ResourceJSON `json:"resources"`
}

type refs struct {
	References []string `json:"references"`
}
