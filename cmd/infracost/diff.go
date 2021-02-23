package main

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/urfave/cli/v2"
)

func diffCmd(cfg *config.Config) *cli.Command {
	flags := make([]cli.Flag, 0)

	flags = append(flags, removeFlag(runInputFlags, "terraform-use-state")...)
	flags = append(flags, removeFlag(runOutputFlags, "format")...)

	return &cli.Command{
		Name:  "diff",
		Usage: "Generates a diff view of costs",
		UsageText: `infracost [global options] diff [command options] [arguments...]

USAGE METHODS:
	# 1. Use terraform directory with any required terraform flags
	infracost diff --terraform-dir /path/to/code --terraform-plan-flags "-var-file=myvars.tfvars"

	# 2. Use terraform state file
	infracost diff --terraform-dir /path/to/code --terraform-use-state

	# 3. Use terraform plan JSON
	terraform plan -out plan.save .
	terraform show -json plan.save > plan.json
	infracost diff--terraform-json-file /path/to/plan.json

	# 4. Use terraform plan file, relative to terraform-dir
	terraform plan -out plan.save .
	infracost diff --terraform-dir /path/to/code --terraform-plan-file plan.save`,
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

			cfg.Outputs = []*config.Output{
				{
					Format: "diff",
				},
			}
			return runMain(cfg)
		},
	}
}

func removeFlag(flags []cli.Flag, nameToRemove string) []cli.Flag {
	l := make([]cli.Flag, 0)

	for _, f := range flags {
		if len(f.Names()) > 0 && f.Names()[0] != nameToRemove {
			l = append(l, f)
		}
	}

	return l
}
