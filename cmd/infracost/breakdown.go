package main

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/urfave/cli/v2"
)

var deprecatedBreakdownFlags = []cli.Flag{
	&cli.StringFlag{
		Name:      "tfjson",
		Usage:     "Path to Terraform plan JSON file",
		TakesFile: true,
		Hidden:    true,
	},
	&cli.StringFlag{
		Name:      "tfplan",
		Usage:     "Path to Terraform plan file relative to 'tfdir'",
		TakesFile: true,
		Hidden:    true,
	},
	&cli.BoolFlag{
		Name:   "use-tfstate",
		Usage:  "Use Terraform state instead of generating a plan",
		Value:  false,
		Hidden: true,
	},
	&cli.StringFlag{
		Name:        "tfdir",
		Usage:       "Path to the Terraform code directory",
		TakesFile:   true,
		DefaultText: "current working directory",
		Hidden:      true,
	},
	&cli.StringFlag{
		Name:   "tfflags",
		Usage:  "Flags to pass to the 'terraform plan' command",
		Hidden: true,
	},
	&cli.StringFlag{
		Name:    "output",
		Aliases: []string{"o"},
		Usage:   "Output format: json, table, html",
		Value:   "table",
		Hidden:  true,
	},
}

func breakdownCmd(cfg *config.Config) *cli.Command {
	flags := make([]cli.Flag, 0)

	flags = append(flags, deprecatedBreakdownFlags...)
	flags = append(flags, runInputFlags...)
	flags = append(flags, runOutputFlags...)

	return &cli.Command{
		Name:  "breakdown",
		Usage: "Generates a full breakdown of costs",
		UsageText: `infracost [global options] breakdown [command options] [arguments...]

USAGE METHODS:
	# 1. Use terraform directory with any required terraform flags
	infracost breakdown --terraform-dir /path/to/code --terraform-plan-flags "-var-file=myvars.tfvars"

	# 2. Use terraform state file
	infracost breakdown --terraform-dir /path/to/code --terraform-use-state

	# 3. Use terraform plan JSON
	terraform plan -out plan.save .
	terraform show -json plan.save > plan.json
	infracost breakdown--terraform-json-file /path/to/plan.json

	# 4. Use terraform plan file, relative to terraform-dir
	terraform plan -out plan.save .
	infracost breakdown --terraform-dir /path/to/code --terraform-plan-file plan.save`,
		Flags: flags,
		Action: func(c *cli.Context) error {
			handleDeprecatedEnvVars(c, deprecatedEnvVarMapping)
			handleDeprecatedFlags(c, deprecatedFlagsMapping)

			if err := checkAPIKey(cfg.APIKey, cfg.PricingAPIEndpoint, cfg.DefaultPricingAPIEndpoint); err != nil {
				return err
			}

			err := loadRunFlags(cfg, c)
			if err != nil {
				return err
			}

			err = checkRunConfig(cfg)
			if err != nil {
				usageError(c, err.Error())
			}

			return runMain(cfg)
		},
	}
}
