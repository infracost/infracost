package cloudformation

import (
	"github.com/awslabs/goformation/v4"
	"github.com/pkg/errors"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
)

type TemplateProvider struct {
	ctx                  *config.ProjectContext
	Path                 string
	includePastResources bool
}

func NewTemplateProvider(ctx *config.ProjectContext, includePastResources bool) schema.Provider {
	return &TemplateProvider{
		ctx:                  ctx,
		Path:                 ctx.ProjectConfig.Path,
		includePastResources: includePastResources,
	}
}

func (p *TemplateProvider) Type() string {
	return "cloudformation"
}

func (p *TemplateProvider) DisplayType() string {
	return "CloudFormation"
}

func (p *TemplateProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	// no op
}

func (p *TemplateProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	template, err := goformation.Open(p.Path)
	if err != nil {
		return []*schema.Project{}, errors.Wrap(err, "Error reading CloudFormation template file")
	}

	metadata := config.DetectProjectMetadata(p.ctx.ProjectConfig.Path)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)
	name := p.ctx.ProjectConfig.Name
	if name == "" {
		name = metadata.GenerateProjectName(p.ctx.RunContext.VCSMetadata.Remote, p.ctx.RunContext.IsCloudEnabled())
	}

	project := schema.NewProject(name, metadata)
	parser := NewParser(p.ctx)
	pastResources, resources, err := parser.parseTemplate(template, usage)
	if err != nil {
		return []*schema.Project{project}, errors.Wrap(err, "Error parsing CloudFormation template file")
	}

	project.PastResources = pastResources
	project.Resources = resources

	if !p.includePastResources {
		project.PastResources = nil
	}

	return []*schema.Project{project}, nil
}
