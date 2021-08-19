package terraform

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
)

var defaultTerragruntBinary = "terragrunt"
var minTerragruntVersion = "v0.31.0"

type TerragruntProvider struct {
	*DirProvider
	Path string
}

func NewTerragruntProvider(ctx *config.ProjectContext) schema.Provider {
	dirProvider := NewDirProvider(ctx).(*DirProvider)

	terragruntBinary := ctx.ProjectConfig.TerraformBinary
	if terragruntBinary == "" {
		terragruntBinary = defaultTerragruntBinary
	}

	dirProvider.TerraformBinary = terragruntBinary
	dirProvider.IsTerragrunt = true

	return &TerragruntProvider{
		DirProvider: dirProvider,
		Path:        ctx.ProjectConfig.Path,
	}
}

func (p *TerragruntProvider) Type() string {
	return "terragrunt"
}

func (p *TerragruntProvider) DisplayType() string {
	return "Terragrunt directory"
}

func (p *TerragruntProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	return p.DirProvider.LoadResources(usage)
}
