package k8s

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
)

var (
	resourceFuncs = map[string]ResourceFunc{
		"v1.Deployment": NewDeploymentResource,
	}
)

type ResourceFunc func(object runtime.Object) (*schema.Resource, error)

type ManifestProvider struct {
	Path          string
	ResourceFuncs map[string]ResourceFunc
	Ctx           *config.ProjectContext
}

func NewManifestProvider(path string, ctx *config.ProjectContext) *ManifestProvider {
	return &ManifestProvider{
		Path:          path,
		ResourceFuncs: resourceFuncs,
		Ctx:           ctx,
	}
}

func (p *ManifestProvider) Type() string {
	return "k8s_dir"
}

func (p *ManifestProvider) DisplayType() string {
	return "K8s Directory"
}

func (p *ManifestProvider) AddMetadata(metadata *schema.ProjectMetadata) {}

func (p *ManifestProvider) LoadResources(usageMap schema.UsageMap) ([]*schema.Project, error) {
	info, err := os.Stat(p.Path)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return p.loadProjectsFromDir(usageMap)
	}

	return p.loadProjectsFromFile(usageMap)
}

func (p *ManifestProvider) readFileResources(path string, usageMap schema.UsageMap) ([]*schema.PartialResource, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var resources []*schema.PartialResource
	decoder := scheme.Codecs.UniversalDeserializer()
	reader := yaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(data)))
	for {
		b, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		obj, _, err := decoder.Decode(b, nil, nil)
		if err != nil {
			return nil, err
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

	return resources, nil
}

func (p *ManifestProvider) loadProjectsFromDir(usageMap schema.UsageMap) ([]*schema.Project, error) {
	files, err := os.ReadDir(p.Path)
	if err != nil {
		return nil, err
	}

	var resources []*schema.PartialResource
	for _, file := range files {
		if file.IsDir() || !IsYMLFile(file) {
			continue
		}

		fileResources, err := p.readFileResources(filepath.Join(p.Path, file.Name()), usageMap)
		if err != nil {
			continue
		}

		resources = append(resources, fileResources...)
	}

	return []*schema.Project{
		p.newProject(resources),
	}, nil
}

func (p *ManifestProvider) loadProjectsFromFile(usageMap schema.UsageMap) ([]*schema.Project, error) {
	fileResources, err := p.readFileResources(p.Path, usageMap)
	if err != nil {
		return nil, err
	}
	return []*schema.Project{
		p.newProject(fileResources),
	}, nil
}

func (p *ManifestProvider) newProject(resources []*schema.PartialResource) *schema.Project {
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

type FileInfo interface {
	Name() string
	IsDir() bool
}

func IsYMLFile(d FileInfo) bool {
	if d.IsDir() {
		return false
	}

	ext := filepath.Ext(d.Name())
	return ext == ".yaml" || ext == ".yml"
}
