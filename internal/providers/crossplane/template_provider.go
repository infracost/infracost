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

// TemplateProvider ...
type TemplateProvider struct {
	ctx  *config.ProjectContext
	Path string
}

// NewTemplateProvider ...
func NewTemplateProvider(ctx *config.ProjectContext) schema.Provider {
	return &TemplateProvider{
		ctx:  ctx,
		Path: ctx.ProjectConfig.Path,
	}
}

// Type returns provider type
func (p *TemplateProvider) Type() string {
	return "crossplane_teplate_file"
}

// DisplayType returns provider display type
func (p *TemplateProvider) DisplayType() string {
	return "Crossplane template file"
}

// AddMetadata ...
func (p *TemplateProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	// no op
}

// LoadResources loads past and current resources
func (p *TemplateProvider) LoadResources(project *schema.Project, usage map[string]*schema.UsageData) error {

	data, err := readCrossPlaneTemplate(p.Path)
	if err != nil {
		log.Error(err)
		return errors.Wrap(err, "Error reading CrossPlane template file")
	}

	parser := NewParser(p.ctx)

	for _, bytes := range data {
		pastResources, resources, err := parser.parseJSON(bytes, usage)
		if err != nil {
			return errors.Wrap(err, "Error parsing CrossPlane template file")
		}
		if len(pastResources) > 0 {
			project.PastResources = append(project.PastResources, pastResources...)
		}
		if len(resources) > 0 {
			project.Resources = append(project.Resources, resources...)
		}
	}

	return nil
}

// IsCrossPlaneTemplateProvider ...
func IsCrossPlaneTemplateProvider(path string) bool {
	template := map[string]interface{}{}
	data, err := readCrossPlaneTemplate(path)
	if err != nil {
		log.Error(err)
		return false
	}
	for _, bytes := range data {
		if err = json.Unmarshal(bytes, &template); err != nil {
			log.Error(err)
			return false
		}
		if err == nil {
			if apiVersionObj, ok := template["apiVersion"]; ok {
				if apiVersion, ok := apiVersionObj.(string); ok && strings.Contains(apiVersion, "crossplane.io/") {
					return true
				}
			}
		}
	}
	return false
}

func readCrossPlaneTemplate(path string) ([][]byte, error) {
	var bytes [][]byte
	extension := filepath.Ext(path)
	if extension != ".yml" && extension != ".yaml" && extension != ".json" {
		err := errors.New("invalid CrossplaneTemplate template file")
		return nil, err
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if extension == ".yml" || extension == ".yaml" {
		fileAsString := string(data[:])
		sepYamlfiles := strings.Split(fileAsString, "---")
		for _, f := range sepYamlfiles {
			if f == "\n" || f == "" {
				// ignore empty cases
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
