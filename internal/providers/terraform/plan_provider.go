package terraform

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type PlanProvider struct {
	*DirProvider
	Path string
}

func NewPlanProvider(ctx *config.ProjectContext) schema.Provider {
	dirProvider := NewDirProvider(ctx).(*DirProvider)

	return &PlanProvider{
		DirProvider: dirProvider,
		Path:        ctx.ProjectConfig.Path,
	}
}

func (p *PlanProvider) Type() string {
	return "terraform_plan"
}

func (p *PlanProvider) DisplayType() string {
	return "Terraform plan file"
}

func (p *PlanProvider) LoadResources(usage map[string]*schema.UsageData) ([]*schema.Project, error) {
	j, err := p.generatePlanJSON()
	if err != nil {
		return []*schema.Project{}, err
	}

	metadata := config.DetectProjectMetadata(p.ctx)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)
	name := schema.GenerateProjectName(metadata, p.ctx.RunContext.Config.EnableDashboard)

	project := schema.NewProject(name, metadata)
	parser := NewParser(p.ctx)

	pastResources, resources, err := parser.parseJSON(j, usage)
	if err != nil {
		return []*schema.Project{project}, errors.Wrap(err, "Error parsing Terraform JSON")
	}

	project.PastResources = pastResources
	project.Resources = resources

	return []*schema.Project{project}, nil
}

func (p *PlanProvider) generatePlanJSON() ([]byte, error) {
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
			return []byte{}, clierror.NewSanitizedError(errors.New(m), "Could not detect Terraform directory for plan file")
		}
	}

	if p.DirProvider != nil {
		p.DirProvider.Path = dir
	}

	err := p.checks()
	if err != nil {
		return []byte{}, err
	}

	opts, err := p.buildCommandOpts()
	if err != nil {
		return []byte{}, err
	}
	if opts.TerraformConfigFile != "" {
		defer os.Remove(opts.TerraformConfigFile)
	}

	return p.runShow(opts, planPath)
}
