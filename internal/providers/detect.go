package providers

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/awslabs/goformation/v4"
	"github.com/infracost/infracost/internal/providers/cloudformation"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/providers/terraform"
	"github.com/infracost/infracost/internal/schema"
)

func Detect(ctx *config.ProjectContext) (schema.Provider, error) {
	path := ctx.ProjectConfig.Path

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("No such file or directory %s", path)
	}

	if isCloudFormationTemplate(path) {
		return cloudformation.NewTemplateProvider(ctx), nil
	}

	if isTerraformPlanJSON(path) {
		return terraform.NewPlanJSONProvider(ctx), nil
	}

	if isTerraformStateJSON(path) {
		return terraform.NewStateJSONProvider(ctx), nil
	}

	if isTerraformPlan(path) {
		return terraform.NewPlanProvider(ctx), nil
	}

	if isTerragruntDir(path) {
		return terraform.NewTerragruntProvider(ctx), nil
	}

	if isTerraformDir(path) {
		return terraform.NewDirProvider(ctx), nil
	}

	return nil, fmt.Errorf("Could not detect path type for %s", path)
}

func isTerraformPlanJSON(path string) bool {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return false
	}

	var jsonFormat struct {
		FormatVersion string      `json:"format_version"`
		PlannedValues interface{} `json:"planned_values"`
	}

	err = json.Unmarshal(b, &jsonFormat)
	if err != nil {
		return false
	}

	return jsonFormat.FormatVersion != "" && jsonFormat.PlannedValues != nil
}

func isTerraformStateJSON(path string) bool {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return false
	}

	var jsonFormat struct {
		FormatVersion string      `json:"format_version"`
		Values        interface{} `json:"values"`
	}

	err = json.Unmarshal(b, &jsonFormat)
	if err != nil {
		return false
	}

	return jsonFormat.FormatVersion != "" && jsonFormat.Values != nil
}

func isTerraformPlan(path string) bool {
	r, err := zip.OpenReader(path)
	if err != nil {
		return false
	}
	defer r.Close()

	var planFile *zip.File
	for _, file := range r.File {
		if file.Name == "tfplan" {
			planFile = file
			break
		}
	}

	return planFile != nil
}

func isTerragruntDir(path string) bool {
	if val, ok := os.LookupEnv("TERRAGRUNT_CONFIG"); ok {
		if filepath.IsAbs(val) {
			return config.FileExists(val)
		}
		return config.FileExists(filepath.Join(path, val))
	}

	return config.FileExists(filepath.Join(path, "terragrunt.hcl")) || config.FileExists(filepath.Join(path, "terragrunt.hcl.json"))
}

func isTerraformDir(path string) bool {
	return terraform.IsTerraformDir(path)
}

func isCloudFormationTemplate(path string) bool {
	template, err := goformation.Open(path)
	if err != nil {
		return false
	}

	if len(template.Resources) > 0 {
		return true
	}

	return false
}
