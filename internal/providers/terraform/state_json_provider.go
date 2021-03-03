package terraform

import (
	"io/ioutil"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
)

type StateJSONProvider struct {
	Path string
	env  *config.Environment
}

func NewStateJSONProvider(cfg *config.Config, projectCfg *config.Project) schema.Provider {
	return &StateJSONProvider{
		Path: projectCfg.Path,
		env:  cfg.Environment,
	}
}

func (p *StateJSONProvider) Type() string {
	return "terraform_state_json"
}

func (p *StateJSONProvider) DisplayType() string {
	return "Terraform state JSON file"
}

func (p *StateJSONProvider) LoadResources(usage map[string]*schema.UsageData) (*schema.Project, error) {
	var project *schema.Project = schema.NewProject(p.Path)

	j, err := ioutil.ReadFile(p.Path)
	if err != nil {
		return project, errors.Wrap(err, "Error reading Terraform state JSON file")
	}

	parser := NewParser(p.env)

	pastResources, resources, err := parser.parseJSON(j, usage)
	if err != nil {
		return project, errors.Wrap(err, "Error parsing Terraform state JSON file")
	}

	project.PastResources = pastResources
	project.Resources = resources

	return project, nil
}
