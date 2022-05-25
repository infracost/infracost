package providers

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/awslabs/goformation/v4"

	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/hcl"
	"github.com/infracost/infracost/internal/providers/cloudformation"
	"github.com/infracost/infracost/internal/providers/pulumi"
	"github.com/infracost/infracost/internal/providers/terraform"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pulumi/pulumi/pkg/v3/engine"
	"github.com/pulumi/pulumi/pkg/v3/resource/deploy"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/common/diag"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

// ValidationError represents an error that is raised because provider conditions are not met.
// This error is commonly used to show requirements to a user running an Infracost command.
type ValidationError struct {
	err  string
	warn string
}

// Warn returns the ValidationError warning message. A warning highlights a potential issue with runtime
// configuration but a condition that the Provider can proceed with.
//
// Warn can return nil if there are no validation warnings.
func (e ValidationError) Warn() *string {
	if e.warn == "" {
		return nil
	}

	return &e.warn
}

// Error returns ValidationError as a string, implementing the error interface.
func (e *ValidationError) Error() string {
	return e.err
}

func Detect(ctx *config.ProjectContext, includePastResources bool) (schema.Provider, error) {
	path := ctx.ProjectConfig.Path

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("No such file or directory %s", path)
	}

	forceCLI := ctx.ProjectConfig.TerraformForceCLI
	projectType := DetectProjectType(path, forceCLI)

	switch projectType {
	case "terraform_dir":
		h, providerErr := terraform.NewHCLProvider(
			ctx,
			nil,
			hcl.OptionWithSpinner(ctx.RunContext.NewSpinner),
			hcl.OptionWithWarningFunc(ctx.RunContext.NewWarningWriter()),
		)

		if providerErr != nil {
			return nil, providerErr
		}

		if err := validateProjectForHCL(ctx, path); err != nil {
			return h, err
		}

		return h, nil
	case "terragrunt_dir":
		h := terraform.NewTerragruntHCLProvider(ctx, includePastResources)
		if err := validateProjectForHCL(ctx, path); err != nil {
			return h, err
		}

		return h, nil
	case "terraform_plan_json":
		return terraform.NewPlanJSONProvider(ctx, includePastResources), nil
	case "terraform_plan_binary":
		return terraform.NewPlanProvider(ctx, includePastResources), nil
	case "terraform_cli":
		return terraform.NewDirProvider(ctx, includePastResources), nil
	case "terragrunt_cli":
		return terraform.NewTerragruntProvider(ctx, includePastResources), nil
	case "terraform_state_json":
		return terraform.NewStateJSONProvider(ctx, includePastResources), nil
	case "cloudformation":
		return cloudformation.NewTemplateProvider(ctx, includePastResources), nil
	case "pulumi":
		return pulumi.NewPreviewJSONProvider(ctx, includePastResources), nil
	}

	return nil, fmt.Errorf("Could not detect path type for '%s'", path)
}

func validateProjectForHCL(ctx *config.ProjectContext, path string) error {
	if ctx.ProjectConfig.TerraformInitFlags != "" {
		return &ValidationError{
			err: "Flag terraform-init-flags is deprecated and only compatible with --terraform-force-cli.",
		}
	}

	if ctx.ProjectConfig.TerraformPlanFlags != "" {
		return &ValidationError{
			err: "Flag terraform-plan-flags is deprecated and only compatible with --terraform-force-cli. If you want to pass Terraform variables use the --terraform-vars or --terraform-var-file flag.",
		}
	}

	if ctx.ProjectConfig.TerraformUseState {
		return &ValidationError{
			err: "Flag terraform-use-state is deprecated and only compatible with --terraform-force-cli.",
		}
	}

	return nil
}

func DetectProjectType(path string, forceCLI bool) string {
	if isCloudFormationTemplate(path) {
		return "cloudformation"
	}

	if isTerraformPlanJSON(path) {
		return "terraform_plan_json"
	}

	if isTerraformStateJSON(path) {
		return "terraform_state_json"
	}

	if isTerraformPlan(path) {
		return "terraform_plan_binary"
	}

	if isTerragruntNestedDir(path, 5) {
		if forceCLI {
			return "terragrunt_cli"
		}
		return "terragrunt_dir"
	}

	if isPulumiPreviewJSON(path) {
		return "pulumi"
	}

	if forceCLI {
		return "terraform_cli"
	}

	return "terraform_dir"
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

func isPulumiPreviewJSON(path string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	var jsonFormat previewDigest

	err = json.Unmarshal(b, &jsonFormat)
	if err != nil {
		return false
	}

	return jsonFormat.ChangeSummary.HasChanges()
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

// previewDigest is a JSON-serializable overview of a preview operation.
type previewDigest struct {
	// Config contains a map of configuration keys/values used during the preview. Any secrets will be blinded.
	Config map[string]string `json:"config,omitempty"`

	// Steps contains a detailed list of all resource step operations.
	Steps []*previewStep `json:"steps,omitempty"`
	// Diagnostics contains a record of all warnings/errors that took place during the preview. Note that
	// ephemeral and debug messages are omitted from this list, as they are meant for display purposes only.
	Diagnostics []previewDiagnostic `json:"diagnostics,omitempty"`

	// Duration records the amount of time it took to perform the preview.
	Duration time.Duration `json:"duration,omitempty"`
	// ChangeSummary contains a map of count per operation (create, update, etc).
	ChangeSummary engine.ResourceChanges `json:"changeSummary,omitempty"`
	// MaybeCorrupt indicates whether one or more resources may be corrupt.
	MaybeCorrupt bool `json:"maybeCorrupt,omitempty"`
}

// propertyDiff contains information about the difference in a single property value.
type propertyDiff struct {
	// Kind is the kind of difference.
	Kind string `json:"kind"`
	// InputDiff is true if this is a difference between old and new inputs instead of old state and new inputs.
	InputDiff bool `json:"inputDiff"`
}

// previewStep is a detailed overview of a step the engine intends to take.
type previewStep struct {
	// Op is the kind of operation being performed.
	Op deploy.StepOp `json:"op"`
	// URN is the resource being affected by this operation.
	URN resource.URN `json:"urn"`
	// Provider is the provider that will perform this step.
	Provider string `json:"provider,omitempty"`
	// OldState is the old state for this resource, if appropriate given the operation type.
	OldState *apitype.ResourceV3 `json:"oldState,omitempty"`
	// NewState is the new state for this resource, if appropriate given the operation type.
	NewState *apitype.ResourceV3 `json:"newState,omitempty"`
	// DiffReasons is a list of keys that are causing a diff (for updating steps only).
	DiffReasons []resource.PropertyKey `json:"diffReasons,omitempty"`
	// ReplaceReasons is a list of keys that are causing replacement (for replacement steps only).
	ReplaceReasons []resource.PropertyKey `json:"replaceReasons,omitempty"`
	// DetailedDiff is a structured diff that indicates precise per-property differences.
	DetailedDiff map[string]propertyDiff `json:"detailedDiff"`
}

// previewDiagnostic is a warning or error emitted during the execution of the preview.
type previewDiagnostic struct {
	URN      resource.URN  `json:"urn,omitempty"`
	Prefix   string        `json:"prefix,omitempty"`
	Message  string        `json:"message,omitempty"`
	Severity diag.Severity `json:"severity,omitempty"`
}
