package terraform

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
)

type PlanProvider struct {
	*DirProvider
	Path                 string
	cachedPlanJSON       []byte
	includePastResources bool
}

func NewPlanProvider(ctx *config.ProjectContext, includePastResources bool) schema.Provider {
	dirProvider := NewDirProvider(ctx, includePastResources).(*DirProvider)

	return &PlanProvider{
		DirProvider:          dirProvider,
		Path:                 ctx.ProjectConfig.Path,
		includePastResources: includePastResources,
	}
}

func (p *PlanProvider) Type() string {
	return "terraform_plan_binary"
}

func (p *PlanProvider) DisplayType() string {
	return "Terraform plan binary file"
}

func (p *PlanProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	j, err := p.generatePlanJSON()
	if err != nil {
		return []*schema.Project{}, err
	}

	spinner := ui.NewSpinner("Extracting only cost-related params from terraform", ui.SpinnerOptions{
		EnableLogging: p.ctx.RunContext.Config.IsLogging(),
		NoColor:       p.ctx.RunContext.Config.NoColor,
		Indent:        "  ",
	})
	defer spinner.Fail()

	metadata := config.DetectProjectMetadata(p.ctx.ProjectConfig.Path)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)
	name := schema.GenerateProjectName(metadata, p.ctx.ProjectConfig.Name, p.ctx.RunContext.IsCloudEnabled())

	project := schema.NewProject(name, metadata)
	parser := NewParser(p.ctx, p.includePastResources)

	pastResources, resources, err := parser.parseJSON(j, usage)
	if err != nil {
		return []*schema.Project{project}, errors.Wrap(err, "Error parsing Terraform JSON")
	}

	project.PastResources = pastResources
	project.Resources = resources

	spinner.Success()
	return []*schema.Project{project}, nil
}

func (p *PlanProvider) generatePlanJSON() ([]byte, error) {
	if p.cachedPlanJSON != nil {
		return p.cachedPlanJSON, nil
	}

	dir := filepath.Dir(p.Path)
	planPath := filepath.Base(p.Path)

	if !IsTerraformDir(dir) {
		log.Debugf("%s is not a Terraform directory, checking current working directory", dir)
		dir, err := os.Getwd()
		if err != nil {
			return []byte{}, err
		}
		planPath = p.Path

		if !IsTerraformDir(dir) {
			m := fmt.Sprintf("%s %s.\n%s\n\n%s\n%s\n%s %s",
				"Could not detect Terraform directory for",
				p.Path,
				"Either the current working directory or the plan file's parent directory must be a Terraform directory.",
				"If the above does not work you can generate the plan JSON file with:",
				ui.PrimaryString("terraform show -json tfplan.binary > plan.json"),
				"and then run Infracost with",
				ui.PrimaryString("--path=plan.json"),
			)
			return []byte{}, clierror.NewCLIError(errors.New(m), "Could not detect Terraform directory for plan file")
		}
	}

	err := p.checks()
	if err != nil {
		return []byte{}, err
	}

	opts, err := p.buildCommandOpts(dir)
	if err != nil {
		return []byte{}, err
	}
	if opts.TerraformConfigFile != "" {
		defer os.Remove(opts.TerraformConfigFile)
	}

	spinner := ui.NewSpinner("Running terraform show", p.spinnerOpts)
	defer spinner.Fail()

	j, err := p.runShow(opts, spinner, planPath)
	if err == nil {
		p.cachedPlanJSON = j
	}
	return j, err
}
