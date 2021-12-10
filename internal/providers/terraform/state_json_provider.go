package terraform

import (
	"os"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
)

type StateJSONProvider struct {
	ctx  *config.ProjectContext
	Path string
}

func NewStateJSONProvider(ctx *config.ProjectContext) schema.Provider {
	return &StateJSONProvider{
		ctx:  ctx,
		Path: ctx.ProjectConfig.Path,
	}
}

func (p *StateJSONProvider) Type() string {
	return "terraform_state_json"
}

func (p *StateJSONProvider) DisplayType() string {
	return "Terraform state JSON file"
}

func (p *StateJSONProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	// no op
}

func (p *StateJSONProvider) LoadResources(ctx *config.ProjectContext, usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	j, err := os.ReadFile(p.Path)
	if err != nil {
		return []*schema.Project{}, errors.Wrap(err, "Error reading Terraform state JSON file")
	}

	metadata := schema.DetectProjectMetadata(p.ctx.ProjectConfig.Path)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)
	name := schema.GenerateProjectName(metadata, p.ctx.Config.EnableDashboard)

	project := schema.NewProject(name, metadata)
	parser := NewParser(p.ctx)

	pastResources, resources, err := parser.parseJSON(ctx, j, usage)
	if err != nil {
		return []*schema.Project{project}, errors.Wrap(err, "Error parsing Terraform state JSON file")
	}

	project.PastResources = pastResources
	project.Resources = resources

	return []*schema.Project{project}, nil
}
