package idem

import (
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"os"

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
	return "idem"
}

func (p *TemplateProvider) DisplayType() string {
	return "Idem"
}

func (p *TemplateProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	metadata.ConfigSha = p.ctx.ProjectConfig.ConfigSha
}

func (p *TemplateProvider) LoadResources(usage schema.UsageMap) ([]*schema.Project, error) {
	b, err := os.ReadFile(p.Path)
	if err != nil {
		return []*schema.Project{}, errors.Wrap(err, "Error reading Idem file")
	}
	if !gjson.ValidBytes(b) {
		return []*schema.Project{}, errors.Wrap(err, "invalid JSON")
	}

	parsed := gjson.ParseBytes(b)

	metadata := config.DetectProjectMetadata(p.ctx.ProjectConfig.Path)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)
	name := p.ctx.ProjectConfig.Name
	if name == "" {
		name = metadata.GenerateProjectName(p.ctx.RunContext.VCSMetadata.Remote, p.ctx.RunContext.IsCloudEnabled())
	}

	project := schema.NewProject(name, metadata)
	parser := NewParser(p.ctx, p.includePastResources)
	pastResources, resources, err := parser.parseTemplate(parsed, usage)
	if err != nil {
		return []*schema.Project{project}, errors.Wrap(err, "Error parsing CloudFormation template file")
	}

	project.PartialPastResources = pastResources
	project.PartialResources = resources

	if !p.includePastResources {
		project.PastResources = nil
	}

	return []*schema.Project{project}, nil
}
