package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/infracost/infracost/internal/ui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var deprecatedFlagsMapping = map[string]string{
	"tfjson":               "path",
	"tfplan":               "path",
	"tfdir":                "path",
	"terraform-json-file":  "path",
	"terraform-plan-file":  "path",
	"terraform-dir":        "path",
	"use-tfstate":          "terraform-use-state",
	"tfflags":              "terraform-plan-flags",
	"output":               "format",
	"o":                    "format",
	"pricing-api-endpoint": "",
}

var deprecatedEnvVarMapping = map[string]string{
	"TERRAFORM_BINARY":      "INFRACOST_TERRAFORM_BINARY",
	"TERRAFORM_CLOUD_HOST":  "INFRACOST_TERRAFORM_CLOUD_HOST",
	"TERRAFORM_CLOUD_TOKEN": "INFRACOST_TERRAFORM_CLOUD_TOKEN",
	"SKIP_UPDATE_CHECK":     "INFRACOST_SKIP_UPDATE_CHECK",
}

func handleDeprecatedEnvVars(deprecatedEnvVars map[string]string) {
	hasPrinted := false

	for oldName, newName := range deprecatedEnvVars {
		if val, ok := os.LookupEnv(oldName); ok {
			m := fmt.Sprintf("Environment variable %s is deprecated and will be removed in v0.8.0.", oldName)
			if newName != "" {
				m += fmt.Sprintf(" Please use %s.", newName)
			}

			ui.PrintWarning(m)
			hasPrinted = true

			if _, ok := os.LookupEnv(newName); !ok {
				os.Setenv(newName, val)
			}
		}
	}

	if hasPrinted {
		fmt.Fprintln(os.Stderr, "")
	}
}

func handleDeprecatedFlags(cmd *cobra.Command, deprecatedFlagsMapping map[string]string) {
	hasPrinted := false

	cmd.Flags().Visit(func(flag *pflag.Flag) {
		if newName, ok := deprecatedFlagsMapping[flag.Name]; ok {

			oldNames := []string{fmt.Sprintf("--%s", flag.Name)}
			if flag.Shorthand != "" {
				oldNames = append(oldNames, fmt.Sprintf("-%s", flag.Shorthand))
			}

			m := fmt.Sprintf("Flag %s is deprecated and will be removed in v0.8.0.", strings.Join(oldNames, "/"))
			if newName != "" {
				m += fmt.Sprintf(" Please use --%s.", newName)
			}

			ui.PrintWarning(m)
			hasPrinted = true

			if !cmd.Flags().Changed(newName) {
				err := cmd.Flags().Set(newName, flag.Value.String())
				if err != nil {
					log.Debugf("Error setting flag %s from %s", newName, flag.Name)
				}
			}
		}
	})

	if hasPrinted {
		fmt.Fprintln(os.Stderr, "")
	}
}

func addDeprecatedRunFlags(cmd *cobra.Command) {
	cmd.Flags().String("tfjson", "", "Path to Terraform plan JSON file")
	_ = cmd.Flags().MarkHidden("tfjson")

	cmd.Flags().String("tfplan", "", "Path to Terraform plan file relative to 'terraform-dir'")
	_ = cmd.Flags().MarkHidden("tfplan")

	cmd.Flags().String("tfflags", "", "Flags to pass to the 'terraform plan' command")
	_ = cmd.Flags().MarkHidden("tfflags")

	cmd.Flags().String("tfdir", "", "Path to the Terraform code directory. Defaults to current working directory")
	_ = cmd.Flags().MarkHidden("tfdir")

	cmd.Flags().Bool("use-tfstate", false, "Use Terraform state instead of generating a plan")
	_ = cmd.Flags().MarkHidden("use-tfstate")

	cmd.Flags().StringP("output", "o", "table", "Output format: json, table, html")
	_ = cmd.Flags().MarkHidden("output")

	cmd.Flags().String("pricing-api-endpoint", "", "Specify an alternate Cloud Pricing API URL")
	_ = cmd.Flags().MarkHidden("pricing-api-endpoint")

	cmd.Flags().String("terraform-json-file", "", "Path to Terraform plan JSON file")
	_ = cmd.Flags().MarkHidden("terraform-json-file")

	cmd.Flags().String("terraform-plan-file", "", "Path to Terraform plan file relative to 'terraform-dir'")
	_ = cmd.Flags().MarkHidden("terraform-plan-file")

	cmd.Flags().String("terraform-dir", "", "Path to the Terraform code directory. Defaults to current working directory")
	_ = cmd.Flags().MarkHidden("terraform-dir")
}
