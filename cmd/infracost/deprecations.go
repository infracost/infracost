package main

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var deprecatedFlagsMapping = map[string]string{
	"tfjson":      "terraform-json-file",
	"tfplan":      "terraform-plan-file",
	"use-tfstate": "terraform-use-state",
	"tfdir":       "terraform-dir",
	"tfflags":     "terraform-plan-flags",
	"output":      "format",
	"o":           "format",
}

var deprecatedEnvVarMapping = map[string]string{
	"TERRAFORM_BINARY":      "INFRACOST_TERRAFORM_BINARY",
	"TERRAFORM_CLOUD_HOST":  "INFRACOST_TERRAFORM_CLOUD_HOST",
	"TERRAFORM_CLOUD_TOKEN": "INFRACOST_TERRAFORM_CLOUD_TOKEN",
	"SKIP_UPDATE_CHECK":     "INFRACOST_SKIP_UPDATE_CHECK",
}

func handleDeprecatedEnvVars(c *cli.Context, deprecatedEnvVars map[string]string) {
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

func handleDeprecatedFlags(c *cli.Context, deprecatedFlagsMapping map[string]string) {
	flags := append(c.App.Flags, c.Command.Flags...)

	for _, flag := range flags {
		flagName := flag.Names()[0]
		if !c.IsSet(flagName) {
			continue
		}

		if newName, ok := deprecatedFlagsMapping[flagName]; ok {

			oldNames := make([]string, 0, len(flag.Names()))
			for _, n := range flag.Names() {
				if len(n) == 1 {
					oldNames = append(oldNames, fmt.Sprintf("-%s", n))
				} else {
					oldNames = append(oldNames, fmt.Sprintf("--%s", n))
				}
			}

			m := fmt.Sprintf("Flag %s is deprecated and will be removed in v0.8.0.", strings.Join(oldNames, "/"))
			if newName != "" {
				m += fmt.Sprintf(" Please use --%s.", newName)
			}

			usageWarning(m)

			if !c.IsSet(newName) {
				err := c.Set(newName, c.String(flagName))
				if err != nil {
					log.Debugf("Error setting flag %s from %s", newName, flagName)
				}
			}
		}
	}
}
