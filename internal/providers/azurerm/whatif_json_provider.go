package azurerm

import (
	"os"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
	"github.com/pkg/errors"
)

// TODO: AzureRM doesn't have a concept of a 'Project', needs its own config.ProjectContext object
type WhatifJsonProvider struct {
	ctx                  *config.ProjectContext
	Path                 string
	includePastResources bool
}

func NewWhatifJsonProvider(ctx *config.ProjectContext, includePastResources bool) schema.Provider {
	return &WhatifJsonProvider{
		ctx:                  ctx,
		Path:                 ctx.ProjectConfig.Path,
		includePastResources: includePastResources,
	}
}

func (p *WhatifJsonProvider) Type() string {
	return "azurerm_whatif_json"
}

func (p *WhatifJsonProvider) DisplayType() string {
	return "Azure Resource Manager WhatIf JSON"
}

func (p *WhatifJsonProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	// no op
}

func (p *WhatifJsonProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	spinner := ui.NewSpinner("Extracting only cost-related params from WhatIf", ui.SpinnerOptions{
		EnableLogging: p.ctx.RunContext.Config.IsLogging(),
		NoColor:       p.ctx.RunContext.Config.NoColor,
		Indent:        "  ",
	})
	defer spinner.Fail()

	j, err := os.ReadFile(p.Path)
	if err != nil {
		return []*schema.Project{}, errors.Wrap(err, "Error reading WhatIf result JSON file")
	}

	metadata := config.DetectProjectMetadata(p.ctx.ProjectConfig.Path)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)

	name := p.ctx.ProjectConfig.Name
	if name == "" {
		name = metadata.GenerateProjectName(p.ctx.RunContext.VCSMetadata.Remote, p.ctx.RunContext.IsCloudEnabled())
	}

	// TODO: This should probably do a call to the whatif endpoint for the subscription
	// Then pass the response to the code below
	// For now, pass a whatif result JSON file directly

	project := schema.NewProject(name, metadata)
	parser := NewParser(p.ctx)

	// TODO: pastResources are ??, check what they are in Azure context
	partialPastResources, partialResources, err := parser.parse(j, usage)
	if err != nil {
		return []*schema.Project{}, errors.Wrap(err, "Error parsing WhatIf data")
	}

	project.PartialPastResources = partialPastResources
	project.PartialResources = partialResources

	spinner.Success()

	return []*schema.Project{project}, nil
}
