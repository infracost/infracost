package pulumi

import (
	"encoding/json"
	"os"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/common/display"
	"github.com/tidwall/gjson"
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
	return "pulumi_preview_json"
}

func (p *PreviewJSONProvider) DisplayType() string {
	return "Pulumi preview JSON file"
}

func (p *PreviewJSONProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	// no op
}

func (p *PreviewJSONProvider) LoadResources(usage schema.UsageMap) ([]*schema.Project, error) {
	b, err := os.ReadFile(p.Path)
	if err != nil {
		return []*schema.Project{}, errors.Wrap(err, "Error reading Pulumi preview JSON file")
	}
	var jsonPreviewDigest display.PreviewDigest
	err = json.Unmarshal(b, &jsonPreviewDigest)
	gjsonResult := gjson.ParseBytes(b)

	if err != nil {
		return []*schema.Project{}, errors.Wrap(err, "Error reading Pulumi preview JSON file")
	}

	metadata := config.DetectProjectMetadata(p.ctx.ProjectConfig.Path)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)
	name := metadata.GenerateProjectName(p.ctx.RunContext.VCSMetadata.Remote, p.ctx.RunContext.IsCloudEnabled())

	project := schema.NewProject(name, metadata)
	parser := NewParser(p.ctx)
	pastResources, resources, err := parser.parsePreviewDigest(jsonPreviewDigest, usage, gjsonResult)
	if err != nil {
		return []*schema.Project{project}, errors.Wrap(err, "Error parsing Pulumi preview JSON file")
	}

	project.PartialPastResources = pastResources
	project.PartialResources = resources

	if !p.includePastResources {
		project.PartialPastResources = nil
	}

	return []*schema.Project{project}, nil
}
