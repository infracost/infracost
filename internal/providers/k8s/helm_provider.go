package k8s

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
)

type HelmProvider struct {
	Path          string
	ResourceFuncs map[string]ResourceFunc
	Ctx           *config.ProjectContext
}

func NewHelmProvider(path string, ctx *config.ProjectContext) *HelmProvider {
	return &HelmProvider{
		Path:          path,
		ResourceFuncs: resourceFuncs,
		Ctx:           ctx,
	}
}

func (p *HelmProvider) Type() string {
	return "helm_dir"
}

func (p *HelmProvider) DisplayType() string {
	return "Helm Chart"
}

func (p *HelmProvider) AddMetadata(metadata *schema.ProjectMetadata) {}

func (p *HelmProvider) LoadResources(usageMap schema.UsageMap) ([]*schema.Project, error) {
	chart, err := loader.Load(p.Path)
	if err != nil {
		return nil, err
	}

	chartValues, err := chartutil.ToRenderValues(chart, p.inputValues(), chartutil.ReleaseOptions{
		Name:      "infracost",
		Namespace: "default",
	}, nil)
	if err != nil {
		return nil, err
	}

	renderedTemplates, err := engine.Render(chart, chartValues)
	if err != nil {
		return nil, err
	}

	var resources []*schema.PartialResource

	for _, template := range renderedTemplates {
		obj, _, err := scheme.Codecs.UniversalDeserializer().Decode([]byte(template), nil, nil)
		if err != nil {
			continue
		}

		t := strings.TrimPrefix(fmt.Sprintf("%T", obj), "*")
		f, ok := p.ResourceFuncs[t]
		if !ok {
			continue
		}

		resource, err := f(obj)
		if err != nil {
			continue
		}
		resources = append(resources, &schema.PartialResource{
			ResourceData: &schema.ResourceData{
				Type:         fmt.Sprintf("k8s.%s", t),
				ProviderName: "k8s",
				Address:      resource.Name,
			},
			Resource: resource,
		})
	}

	return []*schema.Project{
		p.newProject(resources),
	}, nil
}

func (p *HelmProvider) inputValues() chartutil.Values {
	inputValues := chartutil.Values{}
	if len(p.Ctx.ProjectConfig.HelmValuesFiles) > 0 {
		for _, file := range p.Ctx.ProjectConfig.HelmValuesFiles {
			values, err := chartutil.ReadValuesFile(file)
			if err != nil {
				logging.Logger.WithError(err).Errorf("could not load values file %s", file)
			}
			inputValues = chartutil.CoalesceTables(inputValues, values)
		}

		return inputValues
	}

	valuesFile := filepath.Join(p.Path, "values.yaml")
	if _, err := os.Stat(valuesFile); err == nil {
		inputValues, err = chartutil.ReadValuesFile(valuesFile)
		if err != nil {

		}
	}

	return inputValues
}

func (p *HelmProvider) newProject(resources []*schema.PartialResource) *schema.Project {
	name := p.Ctx.ProjectConfig.Name
	if name == "" {
		name = p.Path
	}

	return &schema.Project{
		Metadata: &schema.ProjectMetadata{
			Type: "k8s",
		},
		Name:             name,
		PartialResources: resources,
	}
}

func readValuesFile(path string) (chartutil.Values, error) {
	values := chartutil.Values{}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &values)
	if err != nil {
		return nil, err
	}

	return values, nil
}
