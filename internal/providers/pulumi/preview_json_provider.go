package pulumi

import (
	"github.com/awslabs/goformation/v4"
	"github.com/pkg/errors"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
)

type PreviewJSONProvider struct {
	ctx                  *config.ProjectContext
	Path                 string
	includePastResources bool
}

func NewPreviewJSONProvider(ctx *config.ProjectContext, includePastResources bool) schema.Provider {
	return &PreviewJSONProvider{
		ctx:                  ctx,
		Path:                 ctx.ProjectConfig.Path,
		includePastResources: includePastResources,
	}
}

func (p *PreviewJSONProvider) Type() string {
	return "cloudformation"
}

func (p *PreviewJSONProvider) DisplayType() string {
	return "CloudFormation"
}

func (p *PreviewJSONProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	// no op
}

func (p *PreviewJSONProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	template, err := goformation.Open(p.Path)
	if err != nil {
		return []*schema.Project{}, errors.Wrap(err, "Error reading Pulumi Preview JSON file")
	}

	metadata := config.DetectProjectMetadata(p.ctx.ProjectConfig.Path)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)
	name := schema.GenerateProjectName(metadata, p.ctx.RunContext.Config.EnableDashboard)

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
