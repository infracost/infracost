package providers

import (
	"archive/zip"
	"encoding/json"
	"io/ioutil"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/providers/terraform"
	"github.com/infracost/infracost/internal/schema"
)

func Detect(cfg *config.Config, projectCfg *config.Project) schema.Provider {
	if isTerraformPlanJSON(projectCfg.Path) {
		return terraform.NewPlanJSONProvider(cfg, projectCfg)
	}

	if isTerraformPlan(projectCfg.Path) {
		return terraform.NewPlanProvider(cfg, projectCfg)
	}

	if isTerraformDir(projectCfg.Path) {
		return terraform.NewDirProvider(cfg, projectCfg)
	}

	return nil
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

func isTerraformPlanJSON(path string) bool {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return false
	}

	var jsonFormat struct {
		FormatVersion interface{} `json:"format_version"`
		PlannedValues interface{} `json:"planned_values"`
	}

	err = json.Unmarshal(b, &jsonFormat)
	return err == nil
}

func isTerraformDir(path string) bool {
	return terraform.IsTerraformDir(path)
}
