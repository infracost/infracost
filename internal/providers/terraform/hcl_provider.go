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
	modules, err := p.Parser.ParseDirectory()
	if err != nil {
		return nil, err
	}

	sch := p.modulesToPlanJSON(modules)
	b, err := json.Marshal(sch)
	if err != nil {
		return nil, fmt.Errorf("error handling built plan json from hcl %w", err)
	}

	return p.Provider.LoadResourcesFromSrc(usage, b, nil)
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
			RootModule: struct {
				Resources []ResourceData `json:"resources"`
			}{
				Resources: []ResourceData{},
			},
		},
	}

	for _, module := range modules {
		for _, block := range module.Blocks {
			if block.Type() == "resource" {
				r := ResourceJSON{
					Address:       block.FullName(),
					Mode:          "managed",
					Type:          block.TypeLabel(),
					Name:          block.LocalName(),
					ProviderName:  "registry.terraform.io/hashicorp/aws",
					SchemaVersion: 1,
				}

				c := ResourceChangesJSON{
					Address:      block.FullName(),
					Mode:         "managed",
					Type:         block.TypeLabel(),
					Name:         block.LocalName(),
					ProviderName: "registry.terraform.io/hashicorp/aws",
					Change: struct {
						Actions []string               `json:"actions"`
						Before  interface{}            `json:"before"`
						After   map[string]interface{} `json:"after"`
					}{
						Actions: []string{"create"},
					},
				}

				jsonValues := marshalAttributeValues(block.Values())
				marshalBlock(block, jsonValues)

				c.Change.After = jsonValues
				r.Values = jsonValues

				sch.Configuration.RootModule.Resources = append(sch.Configuration.RootModule.Resources, ResourceData{
					Address:     block.FullName(),
					Mode:        "managed",
					Type:        block.TypeLabel(),
					Name:        block.LocalName(),
					Expressions: blockToReferences(block),
				})

				sch.ResourceChanges = append(sch.ResourceChanges, c)
				if !block.HasModuleBlock() {
					sch.PlannedValues.RootModule.Resources = append(sch.PlannedValues.RootModule.Resources, r)
					continue
				}

				sch.PlannedValues.RootModule.ChildModules[0].Resources = append(sch.PlannedValues.RootModule.ChildModules[0].Resources, r)
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
		childValues := marshalAttributeValues(b.Values())
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

func marshalAttributeValues(value cty.Value) map[string]interface{} {
	if value == cty.NilVal || value.IsNull() {
		return nil
	}
	ret := make(map[string]interface{})

	it := value.ElementIterator()
	for it.Next() {
		k, v := it.Element()
		vJSON, _ := ctyJson.Marshal(v, v.Type())
		ret[k.AsString()] = json.RawMessage(vJSON)
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
	Variables        interface{} `json:"variables"`
	PlannedValues    struct {
		RootModule PlanRootModule `json:"root_module"`
	} `json:"planned_values"`
	ResourceChanges []ResourceChangesJSON `json:"resource_changes"`
	Configuration   Configuration         `json:"configuration"`
}

type PlanRootModule struct {
	Resources    []ResourceJSON `json:"resources"`
	ChildModules []ChildModule  `json:"child_modules"`
}

type Configuration struct {
	RootModule struct {
		Resources []ResourceData `json:"resources"`
	} `json:"root_module"`
}

type ResourceData struct {
	Address     string                 `json:"address"`
	Mode        string                 `json:"mode"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Expressions map[string]interface{} `json:"expressions"`
}

type ChildModule struct {
	Resources []ResourceJSON `json:"resources"`
}

type refs struct {
	References []string `json:"references"`
}
