package terraform

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/prices"
	"github.com/infracost/infracost/internal/scan"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
)

type PlanJSONProvider struct {
	ctx                  *config.ProjectContext
	Path                 string
	includePastResources bool
	scanner              *scan.TerraformPlanScanner
	logger               *logrus.Entry
}

func NewPlanJSONProvider(ctx *config.ProjectContext, includePastResources bool) *PlanJSONProvider {
	var scanner *scan.TerraformPlanScanner
	if ctx.RunContext.Config.PolicyAPIEndpoint != "" {
		scanner = scan.NewTerraformPlanScanner(ctx.RunContext, ctx.Logger(), prices.GetPrices)
	}

	return &PlanJSONProvider{
		ctx:                  ctx,
		Path:                 ctx.ProjectConfig.Path,
		includePastResources: includePastResources,
		scanner:              scanner,
	}
}

func (p *PlanJSONProvider) Type() string {
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

	project, err := p.LoadResourcesFromSrc(usage, j, spinner)
	if err != nil {
		return nil, err
	}

	return []*schema.Project{project}, nil
}

func (p *PlanJSONProvider) LoadResourcesFromSrc(usage map[string]*schema.UsageData, j []byte, spinner *ui.Spinner) (*schema.Project, error) {
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
		return project, fmt.Errorf("Error parsing Terraform plan JSON file %w", err)
	}

	project.PartialPastResources = partialPastResources
	project.PartialResources = partialResources

	if p.scanner != nil {
		err := p.scanner.ScanPlan(project, j)
		if err != nil {
			p.logger.WithError(err).Debugf("Terraform project %s plan JSON scan failed", project.Name)
		}
	}

	if spinner != nil {
		spinner.Success()
	}

	return project, nil
}
