package cloudformation

import (
	"github.com/awslabs/goformation/v4"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
)

type TemplateJSONProvider struct {
	Path string
	env  *config.Environment
}

func NewTemplateJSONProvider(cfg *config.Config, projectCfg *config.Project) schema.Provider {
	return &TemplateJSONProvider{
		Path: projectCfg.Path,
		env:  cfg.Environment,
	}
}

func (p *TemplateJSONProvider) Type() string {
	return "cloudformation_state_json"
}

func (p *TemplateJSONProvider) DisplayType() string {
	return "Cloudformation state JSON file"
}

func (p *TemplateJSONProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	// no op
}

func (p *TemplateJSONProvider) LoadResources(project *schema.Project, usage map[string]*schema.UsageData) error {
	template, err := goformation.Open(p.Path)
	if err != nil {
		return errors.Wrap(err, "Error reading Cloudformation template file")
	}

	parser := NewParser(p.env)
	pastResources, resources, err := parser.parseTemplate(template, usage)
	if err != nil {
		return errors.Wrap(err, "Error parsing Cloudformation template file")
	}

	project.PastResources = pastResources
	project.Resources = resources

	return nil
}
