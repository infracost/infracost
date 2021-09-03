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

	bytes, err := readCrossPlaneTemplate(p.Path)
	if err != nil {
		log.Error(err)
		return errors.Wrap(err, "Error reading CrossPlane template file")
	}

	parser := NewParser(p.ctx)

	pastResources, resources, err := parser.parseJSON(bytes, usage)
	if err != nil {
		return errors.Wrap(err, "Error parsing CrossPlane template file")
	}

	project.PastResources = pastResources
	project.Resources = resources

	return nil
}

// IsCrossPlaneTemplateProvider ...
func IsCrossPlaneTemplateProvider(path string) bool {
	template := map[string]interface{}{}
	bytes, err := readCrossPlaneTemplate(path)
	if err != nil {
		log.Error(err)
		return false
	}
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
	return false
}

func readCrossPlaneTemplate(path string) ([]byte, error) {
	extension := filepath.Ext(path)
	if extension != ".yml" && extension != ".yaml" && extension != ".json" {
		err := errors.New("invalid CrossplaneTemplate template file")
		return nil, err
	}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if extension == ".yml" || extension == ".yaml" {
		if bytes, err = yamlwrapper.YAMLToJSON(bytes[:]); err != nil {
			log.Error(err)
			return nil, err
		}
	}
	return bytes, nil
}

// ToString ...
func ToString(v interface{}) string {
	b, err := json.Marshal(v)
	log.Error(err)
	return string(b)
}
