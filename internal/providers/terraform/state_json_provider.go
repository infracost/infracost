package terraform

import (
	"os"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"

	"github.com/pkg/errors"
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

func (p *StateJSONProvider) Type() string {
	return "terraform_state_json"
}

func (p *StateJSONProvider) DisplayType() string {
	return "Terraform state JSON file"
}

func (p *StateJSONProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	// no op
}

func (p *StateJSONProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	spinner := ui.NewSpinner("Extracting only cost-related params from terraform", ui.SpinnerOptions{
		EnableLogging: p.ctx.RunContext.Config.IsLogging(),
		NoColor:       p.ctx.RunContext.Config.NoColor,
		Indent:        "  ",
	})
	defer spinner.Fail()

	j, err := os.ReadFile(p.Path)
	if err != nil {
		return []*schema.Project{}, errors.Wrap(err, "Error reading Terraform state JSON file")
	}

	metadata := config.DetectProjectMetadata(p.ctx.ProjectConfig.Path)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)
	name := p.ctx.ProjectConfig.Name
	if name == "" {
		name = metadata.GenerateProjectName(p.ctx.RunContext.VCSMetadata.Remote, p.ctx.RunContext.IsCloudEnabled())
	}

	project := schema.NewProject(name, metadata)
	parser := NewParser(p.ctx, p.includePastResources)

	partialPastResources, partialResources, err := parser.parseJSON(j, usage)
	if err != nil {
		return []*schema.Project{project}, errors.Wrap(err, "Error parsing Terraform state JSON file")
	}

	project.PartialPastResources = partialPastResources
	project.PartialResources = partialResources

	spinner.Success()
	return []*schema.Project{project}, nil
}
