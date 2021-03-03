package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/infracost/infracost/internal/ui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func addRootDeprecatedFlags(cmd *cobra.Command) {
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

	cmd.Flags().String("terraform-json-file", "", "Path to Terraform plan JSON file")
	_ = cmd.Flags().MarkHidden("terraform-json-file")

	cmd.Flags().String("terraform-plan-file", "", "Path to Terraform plan file relative to 'terraform-dir'")
	_ = cmd.Flags().MarkHidden("terraform-plan-file")

	cmd.Flags().String("terraform-dir", "", "Path to the Terraform code directory. Defaults to current working directory")
	_ = cmd.Flags().MarkHidden("terraform-dir")

	cmd.Flags().String("terraform-plan-flags", "", "Flags to pass to the 'terraform plan' command")
	_ = cmd.Flags().MarkHidden("terraform-plan-flags")

	cmd.Flags().Bool("terraform-use-state", false, "Use Terraform state instead of generating a plan")
	_ = cmd.Flags().MarkHidden("terraform-use-state")

	cmd.Flags().String("format", "table", "Output format: json, table, html")
	_ = cmd.Flags().MarkHidden("format")

	cmd.Flags().String("path", "", "Path to the code directory or file")
	_ = cmd.Flags().MarkHidden("path")

	cmd.Flags().String("pricing-api-endpoint", "", "Specify an alternate Cloud Pricing API URL")
	_ = cmd.Flags().MarkHidden("pricing-api-endpoint")

	cmd.Flags().String("config-file", "", "Path to the Infracost config file. Cannot be used with other flags")
	_ = cmd.Flags().MarkHidden("config-file")

	cmd.Flags().String("usage-file", "", "Path to Infracost usage file that specifies values for usage-based resources")
	_ = cmd.Flags().MarkHidden("usage-file")

	cmd.Flags().Bool("show-skipped", false, "Show unsupported resources, some of which might be free. Ignored for JSON outputs")
	_ = cmd.Flags().MarkHidden("show-skipped")
}

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

func processDeprecatedFlags(cmd *cobra.Command) {
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		if newName, ok := deprecatedFlagsMapping[flag.Name]; ok {
			if newName != "" && !cmd.Flags().Changed(newName) {
				err := cmd.Flags().Set(newName, flag.Value.String())
				if err != nil {
					log.Debugf("Error setting flag %s from %s", newName, flag.Name)
				}
			}
		}
	})

	if cmd.Flags().Changed("terraform-plan-file") {
		planPath, _ := cmd.Flags().GetString("terraform-plan-file")

		if cmd.Flags().Changed("terraform-dir") {
			dir, _ := cmd.Flags().GetString("terraform-dir")

			planPath = filepath.Join(dir, planPath)
		}

		err := cmd.Flags().Set("path", planPath)
		if err != nil {
			log.Debugf("Error setting flag path to %s", planPath)
		}
	}

	if cmd.Flags().Changed("tfplan") {
		planPath, _ := cmd.Flags().GetString("tfplan")

		if cmd.Flags().Changed("tfdir") {
			dir, _ := cmd.Flags().GetString("tfdir")

			planPath = filepath.Join(dir, planPath)
		}

		err := cmd.Flags().Set("path", planPath)
		if err != nil {
			log.Debugf("Error setting flag path to %s", planPath)
		}
	}

	format, _ := cmd.Flags().GetString("format")
	if format == "table" || format == "" {
		err := cmd.Flags().Set("format", "table_deprecated")
		if err != nil {
			log.Debug("Error setting flag format to table_deprecated")
		}
	}
}

var deprecatedEnvVarMapping = map[string]string{
	"TERRAFORM_BINARY":      "INFRACOST_TERRAFORM_BINARY",
	"TERRAFORM_CLOUD_HOST":  "INFRACOST_TERRAFORM_CLOUD_HOST",
	"TERRAFORM_CLOUD_TOKEN": "INFRACOST_TERRAFORM_CLOUD_TOKEN",
	"SKIP_UPDATE_CHECK":     "INFRACOST_SKIP_UPDATE_CHECK",
}

func processDeprecatedEnvVars() {
	hasPrinted := false

	for oldName, newName := range deprecatedEnvVarMapping {
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
