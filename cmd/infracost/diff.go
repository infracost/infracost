package main

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/ui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func diffCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Show diff of monthly costs between current and planned state",
		Long:  "Show diff of monthly costs between current and planned state",
		Example: `  Use Terraform directory with any required flags:

      infracost diff --path /path/to/code --terraform-plan-flags "-var-file=my.tfvars"

  Use Terraform plan JSON:

      terraform plan -out tfplan.binary
      terraform show -json tfplan.binary > plan.json
      infracost diff --path plan.json`,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkAPIKey(ctx.Config.APIKey, ctx.Config.PricingAPIEndpoint, ctx.Config.DefaultPricingAPIEndpoint); err != nil {
				return err
			}

			err := loadRunFlags(ctx.Config, cmd)
			if err != nil {
				return err
			}

			err = checkRunConfig(cmd.ErrOrStderr(), ctx.Config)
			if err != nil {
				ui.PrintUsage(cmd)
				return err
			}

			err = checkDiffConfig(ctx.Config)
			if err != nil {
				ui.PrintUsage(cmd)
				return err
			}

			ctx.Config.Format = "diff"

			return runMain(cmd, ctx)
		},
	}

	addRunFlags(cmd)

	cmd.Flags().String("out-file", "", "Save output to a file")

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
