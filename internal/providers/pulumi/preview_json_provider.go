package pulumi

import (
	"fmt"
	"os"
	
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"
)

type PreviewJSONProvider struct {
	ctx                 *config.ProjectContext
	Path                string
	includePastResources bool
}

// NewPreviewJSONProvider creates a new provider from a Pulumi preview JSON file
func NewPreviewJSONProvider(ctx *config.ProjectContext, includePastResources bool) *PreviewJSONProvider {
	return &PreviewJSONProvider{
		ctx:                 ctx,
		Path:                ctx.ProjectConfig.Path,
		includePastResources: includePastResources,
	}
}

// Type returns the provider type
func (p *PreviewJSONProvider) Type() string {
	return "pulumi_preview_json"
}

// DisplayType returns the provider display type
func (p *PreviewJSONProvider) DisplayType() string {
	return "Pulumi (preview.json)"
}

// Project returns the project information
func (p *PreviewJSONProvider) LoadResources(usage schema.UsageMap) ([]*schema.Project, error) {
	j, err := os.ReadFile(p.Path)
	if err != nil {
		return []*schema.Project{}, err
	}

	parser := Parser{
		previewJSON: j,
		ctx:         p.ctx,
		usage:       usage,
		includePastResources: p.includePastResources,
	}

	parsedResources, err := parser.parseResources()
	if err != nil {
		log.Warnf("Error parsing Pulumi preview JSON file: %s", err)
		return []*schema.Project{}, fmt.Errorf("Error parsing Pulumi preview JSON file: %w", err)
	}

	project := schema.NewProject(p.ctx.ProjectConfig.Name, p.ctx.RunContext.Config.VCSRepoURL)
	project.Metadata = p.ctx.ProjectConfig.Metadata
	project.Environment = p.ctx.ProjectConfig.Environment
	project.HasDiff = true
	project.Resources = parsedResources

	return []*schema.Project{project}, nil
}