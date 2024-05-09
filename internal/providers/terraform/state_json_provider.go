package terraform

import (
	"os"

	"github.com/pkg/errors"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
)

type StateJSONProvider struct {
	ctx                  *config.ProjectContext
	Path                 string
	includePastResources bool
}

func NewStateJSONProvider(ctx *config.ProjectContext, includePastResources bool) schema.Provider {
	return &StateJSONProvider{
		ctx:                  ctx,
		Path:                 ctx.ProjectConfig.Path,
		includePastResources: includePastResources,
	}
}

func (p *StateJSONProvider) ProjectName() string {
	return config.CleanProjectName(p.ctx.ProjectConfig.Path)
}

func (p *StateJSONProvider) VarFiles() []string {
	return nil
}

func (p *StateJSONProvider) RelativePath() string {
	return p.ctx.ProjectConfig.Path
}

func (p *StateJSONProvider) Context() *config.ProjectContext { return p.ctx }

func (p *StateJSONProvider) Type() string {
	return "terraform_state_json"
}

func (p *StateJSONProvider) DisplayType() string {
	return "Terraform state JSON file"
}

func (p *StateJSONProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	metadata.ConfigSha = p.ctx.ProjectConfig.ConfigSha
}

func (p *StateJSONProvider) LoadResources(usage schema.UsageMap) ([]*schema.Project, error) {
	logging.Logger.Debug().Msg("Extracting only cost-related params from terraform")

	j, err := os.ReadFile(p.Path)
	if err != nil {
		return []*schema.Project{}, errors.Wrap(err, "Error reading Terraform state JSON file")
	}

	metadata := schema.DetectProjectMetadata(p.ctx.ProjectConfig.Path)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)
	name := p.ctx.ProjectConfig.Name
	if name == "" {
		name = metadata.GenerateProjectName(p.ctx.RunContext.VCSMetadata.Remote, p.ctx.RunContext.IsCloudEnabled())
	}

	project := schema.NewProject(name, metadata)
	parser := NewParser(p.ctx, p.includePastResources)

	j, _ = StripSetupTerraformWrapper(j)
	parsedConf, err := parser.parseJSON(j, usage)
	if err != nil {
		return []*schema.Project{project}, errors.Wrap(err, "Error parsing Terraform state JSON file")
	}

	project.AddProviderMetadata(parsedConf.ProviderMetadata)

	project.PartialPastResources = parsedConf.PastResources
	project.PartialResources = parsedConf.CurrentResources

	return []*schema.Project{project}, nil
}
