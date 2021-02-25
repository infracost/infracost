package main

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var deprecatedFlagsMapping = map[string]string{
	"tfjson":               "terraform-json-file",
	"tfplan":               "terraform-plan-file",
	"use-tfstate":          "terraform-use-state",
	"tfdir":                "terraform-dir",
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
	for oldName, newName := range deprecatedEnvVars {
		if val, ok := os.LookupEnv(oldName); ok {
			m := fmt.Sprintf("Environment variable %s is deprecated and will be removed in v0.8.0.", oldName)
			if newName != "" {
				m += fmt.Sprintf(" Please use %s.", newName)
			}

			usageWarning(m)

			if _, ok := os.LookupEnv(newName); !ok {
				os.Setenv(newName, val)
			}
		}
	}
}

func handleDeprecatedFlags(cmd *cobra.Command, deprecatedFlagsMapping map[string]string) {
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

			usageWarning(m)

			if !cmd.Flags().Changed(newName) {
				err := cmd.Flags().Set(newName, flag.Value.String())
				if err != nil {
					log.Debugf("Error setting flag %s from %s", newName, flag.Name)
				}
			}
		}
	})
}
