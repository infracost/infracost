package cloudformation

import (
	"github.com/awslabs/goformation/v7"
	"github.com/pkg/errors"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
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

func (p *TemplateProvider) LoadResources(usage schema.UsageMap) ([]*schema.Project, error) {
	template, err := goformation.Open(p.Path)
	if err != nil {
		return []*schema.Project{}, errors.Wrap(err, "Error reading CloudFormation template file")
	}

	spinner := ui.NewSpinner("Extracting only cost-related params from cloudformation", ui.SpinnerOptions{
		EnableLogging: p.ctx.RunContext.Config.IsLogging(),
		NoColor:       p.ctx.RunContext.Config.NoColor,
		Indent:        "  ",
	})
	defer spinner.Fail()

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

	spinner.Success()
	return []*schema.Project{project}, nil
}
