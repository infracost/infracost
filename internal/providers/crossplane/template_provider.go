package crossplane

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
	yamlwrapper "github.com/sanathkr/yaml"
	log "github.com/sirupsen/logrus"
)

// TemplateProvider handles the loading and parsing of Crossplane templates.
type TemplateProvider struct {
	ctx                  *config.ProjectContext
	Path                 string
	includePastResources bool
}

// NewTemplateProvider creates a new instance of TemplateProvider for Crossplane templates.
func NewTemplateProvider(ctx *config.ProjectContext, includePastResources bool) schema.Provider {
	return &TemplateProvider{
		ctx:                  ctx,
		Path:                 ctx.ProjectConfig.Path,
		includePastResources: includePastResources,
	}
}

// ProjectName returns the cleaned project name derived from the file path.
func (p *TemplateProvider) ProjectName() string {
	return config.CleanProjectName(p.ctx.ProjectConfig.Path)
}

// VarFiles returns a nil slice as Crossplane doesn't use variable files.
func (p *TemplateProvider) VarFiles() []string {
	return nil
}

// Context returns the project context.
func (p *TemplateProvider) Context() *config.ProjectContext {
	return p.ctx
}

// Type returns the type of provider.
func (p *TemplateProvider) Type() string {
	return "crossplane_template_file"
}

// DisplayType returns the display type of the provider.
func (p *TemplateProvider) DisplayType() string {
	return "Crossplane template file"
}

// AddMetadata adds additional metadata to the project.
func (p *TemplateProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	metadata.ConfigSha = p.ctx.ProjectConfig.ConfigSha
}

// RelativePath returns the relative path of the Crossplane template file.
func (p *TemplateProvider) RelativePath() string {
	return p.ctx.ProjectConfig.Path
}

// LoadResources loads and parses the resources from the Crossplane template.
func (p *TemplateProvider) LoadResources(usage schema.UsageMap) ([]*schema.Project, error) {
    data, err := readCrossPlaneTemplate(p.Path)
    if (err != nil) {
        return []*schema.Project{}, errors.Wrap(err, "Error reading CrossPlane template file")
    }

    metadata := schema.DetectProjectMetadata(p.ctx.ProjectConfig.Path)
    metadata.Type = p.Type()
    p.AddMetadata(metadata)
    name := p.ctx.ProjectConfig.Name
    if name == "" {
        name = metadata.GenerateProjectName(p.ctx.RunContext.VCSMetadata.Remote, p.ctx.RunContext.IsCloudEnabled())
    }

    project := schema.NewProject(name, metadata)
    parser := NewParser(p.ctx, p.includePastResources)

    // Convert usage to the expected map[string]*schema.UsageData
    usageData := make(map[string]*schema.UsageData)
    for k, v := range usage.Data() {
        usageData[k] = v
    }

    parsedResources, err := parser.parseTemplates(data, usageData)
    if err != nil {
        return []*schema.Project{project}, errors.Wrap(err, "Error parsing CrossPlane template file")
    }

    // Instead of converting parsedResources to []*schema.Resource,
    // directly append them as PartialResources to the project.
    for _, pr := range parsedResources {
        if pr.PartialResource != nil {
            project.PartialResources = append(project.PartialResources, pr.PartialResource)
        }
    }

    return []*schema.Project{project}, nil
}

// IsCrossPlaneTemplateProvider checks if the provided file is a Crossplane template.
func IsCrossPlaneTemplateProvider(path string) bool {
	data, err := readCrossPlaneTemplate(path)
	if err != nil {
		log.Error(err)
		return false
	}
	for _, bytes := range data {
		template := map[string]interface{}{}
		if err = json.Unmarshal(bytes, &template); err != nil {
			log.Error(err)
			return false
		}
		if apiVersionObj, ok := template["apiVersion"]; ok {
			if apiVersion, ok := apiVersionObj.(string); ok && (strings.Contains(apiVersion, "crossplane.io/") || strings.Contains(apiVersion, "upbound.io/")) {
				return true
			}
		}
	}
	return false
}

// readCrossPlaneTemplate reads and splits a Crossplane template into individual documents.
func readCrossPlaneTemplate(path string) ([][]byte, error) {
	var bytes [][]byte
	extension := filepath.Ext(path)
	if extension != ".yml" && extension != ".yaml" && extension != ".json" {
		err := errors.New("invalid Crossplane template file")
		return nil, err
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if extension == ".yml" || extension == ".yaml" {
		fileAsString := string(data)
		sepYamlfiles := strings.Split(fileAsString, "---")
		for _, f := range sepYamlfiles {
			if f == "\n" || f == "" {
				continue
			}
			b, err := yamlwrapper.YAMLToJSON([]byte(f))
			if err != nil {
				log.Error(err)
				return nil, err
			}
			bytes = append(bytes, b)
		}
	} else {
		bytes = [][]byte{data}
	}
	return bytes, nil
}
