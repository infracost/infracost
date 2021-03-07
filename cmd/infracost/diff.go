package main

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/ui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func diffCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Show diff of monthly costs between current and planned state",
		Long:  "Show diff of monthly costs between current and planned state",
		Example: `  Use Terraform directory with any required Terraform flags:

      infracost diff --path /path/to/code --terraform-plan-flags "-var-file=my.tfvars"

  Use Terraform plan JSON:

      terraform plan -out tfplan.binary .
      terraform show -json tfplan.binary > plan.json
      infracost diff --path plan.json`,
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

			err = checkDiffConfig(cfg)
			if err != nil {
				ui.PrintUsageErrorAndExit(cmd, err.Error())
			}

			cfg.Format = "diff"

			return runMain(cmd, cfg)
		},
	}

	addRunFlags(cmd)

	return cmd
}

func checkDiffConfig(cfg *config.Config) error {
	for _, projectConfig := range cfg.Projects {
		if projectConfig.TerraformUseState {
			return errors.New("terraform_use_state cannot be used with `infracost diff` as the Terraform state only contains the current state")
		}
	}

	return nil
}
