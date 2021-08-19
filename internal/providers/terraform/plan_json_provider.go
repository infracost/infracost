package terraform

import (
	"io/ioutil"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
)

type PlanJSONProvider struct {
	ctx  *config.ProjectContext
	Path string
}

func NewPlanJSONProvider(ctx *config.ProjectContext) schema.Provider {
	return &PlanJSONProvider{
		ctx:  ctx,
		Path: ctx.ProjectConfig.Path,
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

func (p *PlanJSONProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	j, err := ioutil.ReadFile(p.Path)
	if err != nil {
		return []*schema.Project{}, errors.Wrap(err, "Error reading Terraform plan JSON file")
	}

	metadata := config.DetectProjectMetadata(p.ctx)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)
	name := schema.GenerateProjectName(metadata, p.ctx.RunContext.Config.EnableDashboard)

	project := schema.NewProject(name, metadata)
	parser := NewParser(p.ctx)

	pastResources, resources, err := parser.parseJSON(j, usage)
	if err != nil {
		return []*schema.Project{project}, errors.Wrap(err, "Error parsing Terraform plan JSON file")
	}

	project.PastResources = pastResources
	project.Resources = resources

	return []*schema.Project{project}, nil
}
