package azurerm

import (
	"encoding/json"
	"os"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

type WhatifJsonProvider struct {
	ctx  *config.ProjectContext
	Path string
}

func NewWhatifJsonProvider(ctx *config.ProjectContext) schema.Provider {
	return &WhatifJsonProvider{
		ctx:  ctx,
		Path: ctx.ProjectConfig.Path,
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
	// TODO: This should probably do a call to the whatif endpoint for the subscription
	// Then pass the response to the code below
	// For now, pass the whatif result directly

	b, err := os.ReadFile(p.Path)
	if err != nil {
		return []*schema.Project{}, errors.Wrap(err, "Error reading WhatIf operation result")
	}

	var whatif WhatIf
	err = json.Unmarshal(b, &whatif)
	gjsonResult := gjson.ParseBytes(b)

	if err != nil {
		return []*schema.Project{}, errors.Wrap(err, "Error reading WhatIf operation result")
	}

	metadata := config.DetectProjectMetadata(p.ctx.ProjectConfig.Path)
	metadata.Type = p.Type()

	p.AddMetadata(metadata)
	name := metadata.GenerateProjectName(p.ctx.RunContext.VCSMetadata.Remote, p.ctx.RunContext.IsCloudEnabled())

	project := schema.NewProject(name, metadata)
	parser := NewParser(p.ctx)

	parser.parse

}
