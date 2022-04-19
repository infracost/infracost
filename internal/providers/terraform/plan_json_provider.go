package terraform

import (
	"fmt"
	"os"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
)

type PlanJSONProvider struct {
	ctx          *config.ProjectContext
	Path         string
	providerType string
	parser       *singleProjectParser
}

func NewPlanJSONProvider(ctx *config.ProjectContext, snapshot bool) *PlanJSONProvider {
	p := &PlanJSONProvider{
		ctx:  ctx,
		Path: ctx.ProjectConfig.Path,
	}

	p.parser = newSingleProjectParser(ctx.ProjectConfig.Path, ctx, addProviderTypeMetadata(p), useSnapshot(snapshot))
	return p
}

func (p *PlanJSONProvider) Type() string {
	if p.providerType != "" {
		return p.providerType
	}

	return "terraform_plan_json"
}

func (p *PlanJSONProvider) DisplayType() string {
	return "Terraform plan JSON file"
}

func (p *PlanJSONProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	// TerraformWorkspace isn't used to load resources but we still pass it
	// on so it appears in the project name of the output
	metadata.TerraformWorkspace = p.ctx.ProjectConfig.TerraformWorkspace
}

func (p *PlanJSONProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	spinner := ui.NewSpinner("Extracting only cost-related params from terraform", ui.SpinnerOptions{
		EnableLogging: p.ctx.RunContext.Config.IsLogging(),
		NoColor:       p.ctx.RunContext.Config.NoColor,
		Indent:        "  ",
	})
	defer spinner.Fail()

	j, err := os.ReadFile(p.Path)
	if err != nil {
		return []*schema.Project{}, fmt.Errorf("Error reading Terraform plan JSON file %w", err)
	}

	return p.LoadResourcesFromSrc(usage, j, spinner)
}

func (p *PlanJSONProvider) LoadResourcesFromSrc(usage map[string]*schema.UsageData, j []byte, spinner *ui.Spinner) ([]*schema.Project, error) {
	project, err := p.parser.parseJSON(j, usage)
	if err != nil {
		return []*schema.Project{project}, fmt.Errorf("Error parsing Terraform plan JSON file %w", err)
	}

	if spinner != nil {
		spinner.Success()
	}

	return []*schema.Project{project}, nil
}
