package arm

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

type TemplateProvider struct {
	ctx                  *config.ProjectContext
	Path                 string
	includePastResources bool
	content              Content
}

type Content struct {
	FileContents map[string]FileContent
	MergedBytes  []byte
}

type FileContent struct {
	Schema         string                 `json:"$schema"`
	Parameters     map[string]interface{} `json:"parameters"`
	Variables      map[string]interface{} `json:"variables"`
	ContentVersion string                 `json:"contentVersion"`
	Resources      []interface{}          `json:"resources"`
}

func NewTemplateProvider(ctx *config.ProjectContext, includePastResources bool, path string) *TemplateProvider {
	return &TemplateProvider{
		ctx:                  ctx,
		Path:                 path,
		includePastResources: includePastResources,
		content:              Content{FileContents: map[string]FileContent{}},
	}
}

func (p *TemplateProvider) Type() string {
	return "arm"
}
func (p *TemplateProvider) Context() *config.ProjectContext { return p.ctx }

func (p *TemplateProvider) DisplayType() string {
	return "Azure Resource Manager"
}

func (p *TemplateProvider) AddMetadata(metadata *schema.ProjectMetadata) {
	// no op
}

func (p *TemplateProvider) ProjectName() string {
	return config.CleanProjectName(p.ctx.ProjectConfig.Path)
}

func (p *TemplateProvider) RelativePath() string {
	return p.ctx.ProjectConfig.Path
}

func (p *TemplateProvider) VarFiles() []string {
	return nil
}

func (p *TemplateProvider) LoadResources(usage schema.UsageMap) ([]*schema.Project, error) {

	logging.Logger.Debug().Msg("Extracting only cost-related params from arm template")

	rootPath := p.ctx.ProjectConfig.Path
	if rootPath == "" {
		log.Fatal("Root path is not provided")
	}

	projects := make([]*schema.Project, 0)

	// Merge all the resources from the files in the directory
	p.MergeFileResources(p.Path)

	p.content.MergeBytes()

	project, _ := p.loadProject(p.Path, usage)
	projects = append(projects, project)

	return projects, nil

}

func (p *TemplateProvider) loadProject(filePath string, usage schema.UsageMap) (*schema.Project, error) {

	metadata := schema.DetectProjectMetadata(filePath)
	metadata.Type = p.Type()
	p.AddMetadata(metadata)
	name := p.ctx.ProjectConfig.Name
	if name == "" {
		name = metadata.GenerateProjectName(p.ctx.RunContext.VCSMetadata.Remote, p.ctx.RunContext.IsCloudEnabled())
	}

	project := schema.NewProject(name, metadata)
	p.parseFiles(project, usage)
	p.content.MergedBytes = nil
	return project, nil
}

func (p *TemplateProvider) parseFiles(project *schema.Project, usage schema.UsageMap) {
	parser := NewParser(p.ctx, p.includePastResources)
	content := gjson.ParseBytes(p.content.MergedBytes)
	resources, err := parser.ParseJSON(content, usage)
	if err != nil {
		log.Fatal(err, "Error parsing ARM template JSON")
	}

	for _, res := range resources {
		project.PartialResources = append(project.PartialResources, res.PartialResource)
	}

}

func (p *TemplateProvider) LoadFileContent(filePath string) {

	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Store the file content in the content struct
	var content FileContent
	if err = json.Unmarshal(data, &content); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	// If it is not an ARM template, return
	if !IsARMTemplate(content) {
		return
	}
	p.content.FileContents[filePath] = content

}

func (p *TemplateProvider) MergeFileResources(dirPath string) {

	// If the path is a file, load the file resources
	if strings.HasSuffix(dirPath, ".json") {
		p.LoadFileContent(dirPath)
		return

	}
	// If the path is a directory, load all the file resources in the directory that have a .json extension
	fileInfos, _ := os.ReadDir(dirPath)
	for _, info := range fileInfos {

		if info.IsDir() {
			continue
		}

		name := info.Name()
		filePath := dirPath + "/" + name

		if !strings.HasSuffix(name, ".json") {
			continue
		}
		p.LoadFileContent(filePath)

	}

}

func (c *Content) MergeBytes() {
	var resources []interface{}
	for _, content := range c.FileContents {
		resources = append(resources, content.Resources...)
	}

	mergedBytes, err := json.Marshal(resources)
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	c.MergedBytes = mergedBytes
}

func IsARMTemplate(content FileContent) bool {
	/*
		The schema property is the location of the JavaScript Object Notation (JSON) schema file that describes the version of the template language.
		Since it is a required property in an ARM Template, then it will be used to detect whether the file is an ARM Template or not.

		For more information, see: https://learn.microsoft.com/en-us/azure/azure-resource-manager/templates/syntax
	*/
	if content.Schema == "" {
		return false
	}

	schemaPattern := "^https://schema\\.management\\.azure\\.com/schemas/\\d{4}-\\d{2}-\\d{2}/(tenant|managementGroup|subscription)?deploymentTemplate\\.json#$"
	matched, err := regexp.Match(schemaPattern, []byte(content.Schema))
	if err != nil {
		return false
	}

	// Another way to check if the file is an ARM template is to check if the contentVersion and resources properties are present, since they are required in an ARM template
	return matched && content.ContentVersion != "" && content.Resources != nil
}
