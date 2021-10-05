package cloudformation

import (
	"github.com/awslabs/goformation/v4"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
)

type TemplateProvider struct {
	ctx  *config.ProjectContext
	Path string
}

func NewTemplateProvider(ctx *config.ProjectContext) schema.Provider {
	return &TemplateProvider{
		ctx:  ctx,
		Path: ctx.ProjectConfig.Path,
	}
}

func (p *TemplateProvider) Type() string {
	return "cloudformation_state_json"
}

func (p *TemplateProvider) DisplayType() string {
	return "Cloudformation state JSON file"
}

func (p *TemplateProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	// no op
}

func (p *TemplateProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	template, err := goformation.Open(p.Path)
	if err != nil {
		return []*schema.Project{}, errors.Wrap(err, "Error reading Cloudformation template file")
	}

	metadata := config.DetectProjectMetadata(p.ctx.ProjectConfig.Path)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)
	name := schema.GenerateProjectName(metadata, p.ctx.RunContext.Config.EnableDashboard)

	project := schema.NewProject(name, metadata)
	parser := NewParser(p.ctx)
	pastResources, resources, err := parser.parseTemplate(template, usage)
	if err != nil {
		return []*schema.Project{project}, errors.Wrap(err, "Error parsing Cloudformation template file")
	}

	project.PastResources = pastResources
	project.Resources = resources

	return []*schema.Project{project}, nil
}
