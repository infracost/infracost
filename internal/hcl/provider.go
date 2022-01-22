package hcl

import (
	"encoding/json"
	"fmt"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/hcl/convert"
	"github.com/infracost/infracost/internal/hcl/parser"
	"github.com/infracost/infracost/internal/providers/terraform"
	"github.com/infracost/infracost/internal/schema"
)

type DirProvider struct {
	Ctx      *config.ProjectContext
	Provider *terraform.PlanJSONProvider
}

func (p DirProvider) Type() string {
	return "hcl_provider"
}

func (p DirProvider) DisplayType() string {
	return "HCL Provider"
}

func (p DirProvider) AddMetadata(metadata *schema.ProjectMetadata) {
}

func (p DirProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	option, err := parser.TfVarsToOption(p.Ctx.ProjectConfig.TFVarFiles...)
	if err != nil {
		return nil, err
	}

	modules, err := parser.New(p.Ctx.ProjectConfig.Path, option).ParseDirectory()
	if err != nil {
		return nil, err
	}

	sch := convert.ModulesToPlanJSON(modules)
	b, err := json.Marshal(sch)
	if err != nil {
		return nil, fmt.Errorf("error handling built plan json from hcl %w", err)
	}

	return p.Provider.LoadResourcesFromSrc(usage, b)
}
