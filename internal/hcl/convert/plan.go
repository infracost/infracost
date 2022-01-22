package convert

import (
	"encoding/json"

	"github.com/zclconf/go-cty/cty"
	ctyJson "github.com/zclconf/go-cty/cty/json"

	"github.com/infracost/infracost/internal/hcl/block"
)

func ModulesToPlanJSON(modules []block.Module) PlanSchema {
	sch := PlanSchema{
		FormatVersion:    "1.0",
		TerraformVersion: "1.1.0",
		Variables:        nil,
		PlannedValues: struct {
			RootModule struct {
				Resources    []ResourceJSON `json:"resources"`
				ChildModules []ChildModule  `json:"child_modules"`
			} `json:"root_module"`
		}{
			RootModule: struct {
				Resources    []ResourceJSON `json:"resources"`
				ChildModules []ChildModule  `json:"child_modules"`
			}{
				Resources:    []ResourceJSON{},
				ChildModules: []ChildModule{{}},
			},
		},
		ResourceChanges: []ResourceChangesJSON{},
		Configuration:   nil,
	}

	for _, module := range modules {
		for _, block := range module.GetBlocks() {
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

				for _, b := range block.AllBlocks() {
					childValues := marshalAttributeValues(b.Values())

					if v, ok := jsonValues[b.Type()]; ok {
						jsonValues[b.Type()] = append(v.([]interface{}), childValues)
						continue
					}

					jsonValues[b.Type()] = []interface{}{childValues}
				}

				c.Change.After = jsonValues
				r.Values = jsonValues

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
		RootModule struct {
			Resources    []ResourceJSON `json:"resources"`
			ChildModules []ChildModule  `json:"child_modules"`
		} `json:"root_module"`
	} `json:"planned_values"`
	ResourceChanges []ResourceChangesJSON `json:"resource_changes"`
	Configuration   interface{}           `json:"configuration"`
}

type ChildModule struct {
	Resources []ResourceJSON `json:"resources"`
}
