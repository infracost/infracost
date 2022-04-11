package providers

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/awslabs/goformation/v4"
	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/hcl"
	"github.com/infracost/infracost/internal/providers/cloudformation"
	"github.com/infracost/infracost/internal/providers/terraform"
	"github.com/infracost/infracost/internal/schema"
)

// ValidationError represents an error that is raised because provider conditions are not met.
// This error is commonly used to show requirements to as user running an Infracost command.
type ValidationError struct {
	Msg string
}

// Error returns ValidationError as a string, implementing the error interface.
func (e *ValidationError) Error() string {
	return e.Msg
}

func Detect(ctx *config.ProjectContext) (schema.Provider, error) {
	path := ctx.ProjectConfig.Path

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("No such file or directory %s", path)
	}

	if ctx.ProjectConfig.TerraformParseHCL {
		if err := validateProjectForHCL(ctx, path); err != nil {
			return nil, err
		}

		return terraform.NewHCLProvider(ctx, terraform.NewPlanJSONProvider(ctx), hcl.OptionWithSpinner(ctx.RunContext.NewSpinner))
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

	if isTerragruntNestedDir(path, 5) {
		return terraform.NewTerragruntProvider(ctx), nil
	}

	return nil, fmt.Errorf("Could not detect path type for '%s'", path)
}

func validateProjectForHCL(ctx *config.ProjectContext, path string) error {
	if isTerragruntDir(path) || isTerragruntNestedDir(path, 5) {
		return &ValidationError{
			Msg: "Terragrunt projects are currently not supported by Infracost HCL",
		}
	}

	if isCloudFormationTemplate(path) {
		return &ValidationError{
			Msg: "Cloudformation projects are currently not supported by Infracost HCL",
		}
	}

	if isTerraformPlanJSON(path) {
		return &ValidationError{
			Msg: "Path type cannot be a Plan JSON file when using Infracost HCL\n\nTry setting --path to a Terraform directory.",
		}
	}

	if isTerraformStateJSON(path) {
		return &ValidationError{
			Msg: "Path type cannot be a Terraform state file when using Infracost HCL\n\nTry setting --path to a Terraform directory.",
		}
	}

	if ctx.ProjectConfig.TerraformUseState {
		return &ValidationError{
			Msg: "Flags terraform-use-state and terraform-parse-hcl are incompatible\n\nTry running again without --terraform-use-state",
		}
	}

	for _, k := range os.Environ() {
		if strings.HasPrefix(k, "TF_VAR") {
			log.Warnf("Infracost HCL provider does not support use of TF_VARS through env variables, use --terraform-var-file or --terraform-var instead")
			break
		}
	}

	return nil
}

func isTerraformPlanJSON(path string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	var jsonFormat struct {
		FormatVersion string      `json:"format_version"`
		PlannedValues interface{} `json:"planned_values"`
	}

	b, hasWrapper := terraform.StripSetupTerraformWrapper(b)
	if hasWrapper {
		log.Infof("Stripped wrapper output from %s (to make it a valid JSON file) since setup-terraform GitHub Action was used without terraform_wrapper: false", path)
	}

	err = json.Unmarshal(b, &jsonFormat)
	if err != nil {
		return false
	}

	return jsonFormat.FormatVersion != "" && jsonFormat.PlannedValues != nil
}

func isTerraformStateJSON(path string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	var jsonFormat struct {
		FormatVersion string      `json:"format_version"`
		Values        interface{} `json:"values"`
	}

	b, hasWrapper := terraform.StripSetupTerraformWrapper(b)
	if hasWrapper {
		log.Debugf("Stripped setup-terraform wrapper output from %s", path)
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

func isTerragruntNestedDir(path string, maxDepth int) bool {
	if isTerragruntDir(path) {
		return true
	}

	if maxDepth > 0 {
		entries, err := os.ReadDir(path)
		if err == nil {
			for _, entry := range entries {
				name := entry.Name()
				if entry.IsDir() && name != ".infracost" && name != ".terraform" {
					if isTerragruntNestedDir(filepath.Join(path, name), maxDepth-1) {
						return true
					}
				}
			}
		}
	}
	return false
}

func isTerraformDir(path string) bool {
	return terraform.IsTerraformDir(path)
}

// goformation lib is not threadsafe, so we run this check synchronously
// See: https://github.com/awslabs/goformation/issues/363
var cfMux = &sync.Mutex{}

func isCloudFormationTemplate(path string) bool {
	cfMux.Lock()
	defer cfMux.Unlock()

	template, err := goformation.Open(path)
	if err != nil {
		return false
	}

	if len(template.Resources) > 0 {
		return true
	}

	return false
}
