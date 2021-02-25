package main

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/spf13/cobra"
)

func addDeprecatedBreakdownFlags(cmd *cobra.Command) {
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
}

func breakdownCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "breakdown",
		Short: "Generates a full breakdown of costs",
		Long: `Generates a full breakdown of costs

Use terraform directory with any required terraform flags:

  infracost breakdown --terraform-dir /path/to/code --terraform-plan-flags "-var-file=myvars.tfvars"

Use terraform state file:

  infracost breakdown --terraform-dir /path/to/code --terraform-use-state

Use terraform plan JSON:

  terraform plan -out plan.save .
  terraform show -json plan.save > plan.json
  infracost breakdown --terraform-json-file /path/to/plan.json

Use terraform plan file, relative to terraform-dir:

  terraform plan -out plan.save .
  infracost breakdown --terraform-dir /path/to/code --terraform-plan-file plan.save`,
		PreRun: func(cmd *cobra.Command, args []string) {
			handleDeprecatedEnvVars(deprecatedEnvVarMapping)
			handleDeprecatedFlags(cmd, deprecatedFlagsMapping)
		},
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
				usageError(cmd, err.Error())
			}

			return runMain(cfg)
		},
	}

	addDeprecatedBreakdownFlags(cmd)

	addRunInputFlags(cmd)
	addRunOutputFlags(cmd)

	cmd.Flags().Bool("terraform-use-state", false, "Use Terraform state instead of generating a plan")
	cmd.Flags().String("format", "table", "Output format: json, table, html")

	return cmd
}
