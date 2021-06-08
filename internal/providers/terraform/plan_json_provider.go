package terraform

import (
	"io/ioutil"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
)

type PlanJSONProvider struct {
	Path string
	env  *config.Environment
}

func NewPlanJSONProvider(cfg *config.Config, projectCfg *config.Project) schema.Provider {
	return &PlanJSONProvider{
		Path: projectCfg.Path,
		env:  cfg.Environment,
	}
}

func (p *PlanJSONProvider) Type() string {
	return "terraform_plan_json"
}

func (p *PlanJSONProvider) DisplayType() string {
	return "Terraform plan JSON file"
}

func (p *PlanJSONProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	// no op
}

func (p *PlanJSONProvider) LoadResources(project *schema.Project, usage map[string]*schema.UsageData) error {
	j, err := ioutil.ReadFile(p.Path)
	if err != nil {
		return errors.Wrap(err, "Error reading Terraform plan JSON file")
	}

	parser := NewParser(p.env)

	pastResources, resources, err := parser.parseJSON(j, usage)
	if err != nil {
		return errors.Wrap(err, "Error parsing Terraform plan JSON file")
	}

	project.PastResources = pastResources
	project.Resources = resources

	return nil
}
