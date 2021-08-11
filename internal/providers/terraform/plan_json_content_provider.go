package terraform

import (
	"fmt"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
)

type TerraformJsonPlanProvider struct {
	ctx     *config.ProjectContext
	content []byte
}

func (t TerraformJsonPlanProvider) Type() string {
	return "terraform_plan_json_body"
}

func (t TerraformJsonPlanProvider) DisplayType() string {
	return "Terraform plan JSON"
}

func (t TerraformJsonPlanProvider) AddMetadata(_ *schema.ProjectMetadata) {
	// no op
}

func (t TerraformJsonPlanProvider) LoadResources(project *schema.Project, usage map[string]*schema.UsageData) error {
	parser := NewParser(t.ctx)
	pastResources, resources, err := parser.parseJSON(t.content, usage)
	if err != nil {
		return fmt.Errorf("Error parsing Terraform plan JSON file %w", err)
	}
	project.PastResources = pastResources
	project.Resources = resources
	return nil
}
