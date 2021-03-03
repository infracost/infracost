package terraform

import (
	"os"
	"path/filepath"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
)

type PlanProvider struct {
	*DirProvider
	Path string
	env  *config.Environment
}

func NewPlanProvider(cfg *config.Config, projectCfg *config.TerraformProject) schema.Provider {
	dirProvider := NewDirProvider(cfg, projectCfg).(*DirProvider)
	dirProvider.Path = filepath.Dir(projectCfg.Path)

	return &PlanProvider{
		DirProvider: dirProvider,
		Path:        filepath.Base(projectCfg.Path),
		env:         cfg.Environment,
	}
}

func (p *PlanProvider) Type() string {
	return "Terraform plan file"
}

func (p *PlanProvider) LoadResources(usage map[string]*schema.UsageData) (*schema.Project, error) {
	var project *schema.Project = schema.NewProject(p.Path)

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

	return p.runShow(opts, p.Path)
}
