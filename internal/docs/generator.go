package docs

import (
	"os"
	"text/template"

	"github.com/infracost/infracost/internal/providers/terraform"
)

func generateSupportedResourcesDocs(docsTemplatesPath string, outputPath string) error {
	tmpl, err := template.ParseFiles(docsTemplatesPath + "/supported_resources.md")
	if err != nil {
		return err
	}
	f, err := os.Create(outputPath + "/supported_resources.md")
	if err != nil {
		return err
	}
	err = tmpl.Execute(f, terraform.ResourceRegistryMap)
	if err != nil {
		return err
	}
	return nil
}

func GenerateDocs(docsTemplatesPath, outputPath string) error {
	err := os.MkdirAll(outputPath, os.ModePerm)
	if err != nil {
		return err
	}
	err = generateSupportedResourcesDocs(docsTemplatesPath, outputPath)
	if err != nil {
		return err
	}
	return nil
}
