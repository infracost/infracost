package main

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/ui"
	"github.com/spf13/cobra"
)

func diffCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Generate diff view of costs",
		Long: `Generate diff view of costs

Use terraform directory with any required terraform flags:

  infracost diff --path /path/to/code --terraform-plan-flags "-var-file=myvars.tfvars"

Use terraform state file:

  infracost diff --path /path/to/code --terraform-use-state

Use terraform plan JSON:

  terraform plan -out tfplan.binary .
  terraform show -json tfplan.binary > plan.json
  infracost diff --path /path/to/plan.json

Use terraform plan file:

  terraform plan -out tfplan.binary .
  infracost diff --path tfplan.binary`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkAPIKey(cfg.APIKey, cfg.PricingAPIEndpoint, cfg.DefaultPricingAPIEndpoint); err != nil {
				return err
			}

			err := loadRunFlags(cfg, cmd)
			if err != nil {
				return err
			}

			err = checkRunConfig(cfg)
			if err != nil {
				ui.PrintUsageErrorAndExit(cmd, err.Error())
			}

			cfg.Format = "diff"

			return runMain(cfg)
		},
	}

	addRunFlags(cmd)

	return cmd
}
