package cloudformation

import (
	"github.com/awslabs/goformation/v7"
	"github.com/pkg/errors"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
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

func (p *TemplateProvider) ProjectName() string {
	return config.CleanProjectName(p.ctx.ProjectConfig.Path)
}

func (p *TemplateProvider) VarFiles() []string {
	return nil
}

func (p *TemplateProvider) Context() *config.ProjectContext { return p.ctx }

func (p *TemplateProvider) Type() string {
	return "cloudformation"
}

func (p *TemplateProvider) DisplayType() string {
	return "CloudFormation"
}

func (p *TemplateProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	metadata.ConfigSha = p.ctx.ProjectConfig.ConfigSha
}

func (p *TemplateProvider) RelativePath() string {
	return p.ctx.ProjectConfig.Path
}

func (p *TemplateProvider) LoadResources(usage schema.UsageMap) ([]*schema.Project, error) {
	template, err := goformation.Open(p.Path)
	if err != nil {
		return []*schema.Project{}, errors.Wrap(err, "Error reading CloudFormation template file")
	}

	logging.Logger.Debug().Msg("Extracting only cost-related params from cloudformation")

	metadata := schema.DetectProjectMetadata(p.ctx.ProjectConfig.Path)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)
	name := p.ctx.ProjectConfig.Name
	if name == "" {
		name = metadata.GenerateProjectName(p.ctx.RunContext.VCSMetadata.Remote, p.ctx.RunContext.IsCloudEnabled())
	}

	project := schema.NewProject(name, metadata)
	parser := NewParser(p.ctx, p.includePastResources)
	parsed := parser.parseTemplate(template, usage)
	if err != nil {
		return []*schema.Project{project}, errors.Wrap(err, "Error parsing CloudFormation template file")
	}

	for _, item := range parsed {
		project.PartialResources = append(project.PartialResources, item.PartialResource)
	}

	return []*schema.Project{project}, nil
}
