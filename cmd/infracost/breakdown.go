package main

import (
	"fmt"

	"github.com/infracost/infracost/internal/metrics"
	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/ui"
)

func breakdownCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "breakdown",
		Short: "Show breakdown of costs",
		Long:  "Show breakdown of costs",
		Example: `  Use Terraform directory:

      infracost breakdown --path /code --terraform-var-file my.tfvars

  Use Terraform plan JSON:

      terraform plan -out tfplan.binary
      terraform show -json tfplan.binary > plan.json
      infracost breakdown --path plan.json`,
		ValidArgs: []string{"--", "-"},
		RunE: checkAPIKeyIsValid(ctx, func(cmd *cobra.Command, args []string) error {

			timer := metrics.GetTimer("breakdown.total_duration", false).Start()
			defer func() {
				timer.Stop()
				if path := ctx.Config.MetricsPath; path != "" {
					if err := metrics.WriteMetrics(path); err != nil {
						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error writing metrics: %s\n", err)
					}
				}
			}()

			if err := checkAPIKey(ctx.Config.APIKey, ctx.Config.PricingAPIEndpoint, ctx.Config.DefaultPricingAPIEndpoint); err != nil {
				return err
			}

			err := loadRunFlags(ctx.Config, cmd)
			if err != nil {
				return err
			}

			ctx.ContextValues.SetValue("outputFormat", ctx.Config.Format)

			err = checkRunConfig(cmd.ErrOrStderr(), ctx.Config)
			if err != nil {
				ui.PrintUsage(cmd)
				return err
			}

			return runMain(cmd, ctx)
		}),
	}

	addRunFlags(cmd)

	cmd.Flags().String("out-file", "", "Save output to a file, helpful with format flag")
	cmd.Flags().Bool("terraform-use-state", false, "Use Terraform state instead of generating a plan. Applicable with --terraform-force-cli")
	newEnumFlag(cmd, "format", "table", "Output format", []string{"json", "table", "html"})
	cmd.Flags().StringSlice("fields", []string{"monthlyQuantity", "unit", "monthlyCost"}, "Comma separated list of output fields: all,price,monthlyQuantity,unit,hourlyCost,monthlyCost.\nSupported by table and html output formats")

	// This is deprecated and will show a warning if used without --terraform-force-cli
	_ = cmd.Flags().MarkHidden("terraform-use-state")

	return cmd
}
