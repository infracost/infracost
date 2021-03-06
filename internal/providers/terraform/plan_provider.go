package terraform

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type PlanProvider struct {
	*DirProvider
	Path string
	env  *config.Environment
}

func NewPlanProvider(cfg *config.Config, projectCfg *config.Project) schema.Provider {
	dirProvider := NewDirProvider(cfg, projectCfg).(*DirProvider)

	return &PlanProvider{
		DirProvider: dirProvider,
		Path:        projectCfg.Path,
		env:         cfg.Environment,
	}
}

func (p *PlanProvider) Type() string {
	return "terraform_plan"
}

func (p *PlanProvider) DisplayType() string {
	return "Terraform plan file"
}

func (p *PlanProvider) LoadResources(usage map[string]*schema.UsageData) (*schema.Project, error) {
	var project *schema.Project = schema.NewProject(p.Path, map[string]string{})

	j, err := p.generatePlanJSON()
	if err != nil {
		return project, err
	}

	parser := NewParser(p.env)

	pastResources, resources, err := parser.parseJSON(j, usage)
	if err != nil {
		return project, errors.Wrap(err, "Error parsing Terraform JSON")
	}

	project.PastResources = pastResources
	project.Resources = resources

	return project, nil
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
			return []byte{}, fmt.Errorf("%s %s.\n%s\n\n%s\n%s\n%s %s",
				"Could not detect Terraform directory for",
				p.Path,
				"Either the current working directory or the plan file's parent directory must be a Terraform directory.",
				"If the above does not work you can generate the plan JSON file with:",
				ui.PrimaryString("terraform show -json tfplan.binary > /path/to/plan.json"),
				"and then run Infracost with",
				ui.PrimaryString("--path=/path/to/plan.json"),
			)
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
