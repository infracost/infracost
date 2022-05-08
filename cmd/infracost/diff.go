package main

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/ui"
)

func diffCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Show diff of monthly costs between current and planned state",
		Long:  "Show diff of monthly costs between current and planned state",
		Example: `  Use Terraform directory with any required flags:

      infracost breakdown --path /code --format json --out-file infracost-run.json
      # Make some changes to your Terraform project
      infracost diff --path /code --terraform-var-file my.tfvars --compare-to infracost-run.json

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

			return runDiff(cmd, ctx)
		},
	}

	addRunFlags(cmd)

	cmd.Flags().String("compare-to", "", "Path to Infracost JSON file to compare against")
	newEnumFlag(cmd, "format", "diff", "Output format", []string{"json", "diff"})
	cmd.Flags().String("out-file", "", "Save output to a file")

	return cmd
}

func runDiff(cmd *cobra.Command, ctx *config.RunContext) error {
	if len(ctx.Config.Projects) > 0 {
		path := ctx.Config.Projects[0].Path

		// if the path provided is an Infracost JSON we need to run a compare run
		current, err := output.Load(path)
		if err == nil {
			if ctx.Config.CompareTo == "" {
				return errors.New("Passing an Infracost JSON as a --path argument is only valid using the --compare-to flag")
			}

			return runCompare(cmd, ctx, current)
		}
	}

	return runMain(cmd, ctx)
}

func runCompare(cmd *cobra.Command, ctx *config.RunContext, current output.Root) error {
	prior, err := output.Load(ctx.Config.CompareTo)
	if err != nil {
		return fmt.Errorf("Error loading %s used by --compare-to flag. %s", ctx.Config.CompareTo, err)
	}

	combined, err := output.CompareTo(current, prior)
	if err != nil {
		return err
	}

	format, _ := cmd.Flags().GetString("format")
	b, err := output.FormatOutput(strings.ToLower(format), combined, output.Options{
		DashboardEnabled: ctx.Config.EnableDashboard,
		ShowSkipped:      ctx.Config.ShowSkipped,
		NoColor:          ctx.Config.NoColor,
		Fields:           ctx.Config.Fields,
	})
	if err != nil {
		return err
	}

	pricingClient := apiclient.NewPricingAPIClient(ctx)
	err = pricingClient.AddEvent("infracost-run", ctx.EventEnv())
	if err != nil {
		log.Errorf("Error reporting event: %s", err)
	}

	if outFile, _ := cmd.Flags().GetString("out-file"); outFile != "" {
		return saveOutFile(ctx, cmd, outFile, b)
	}

	cmd.Println(string(b))
	return nil
}

func checkDiffConfig(cfg *config.Config) error {
	for _, projectConfig := range cfg.Projects {
		if projectConfig.TerraformUseState {
			return errors.New("terraform_use_state cannot be used with `infracost diff` as the Terraform state only contains the current state")
		}

		if projectConfig.TerraformParseHCL && cfg.CompareTo == "" {
			return errors.New("Use `infracost diff --path /code --terraform-parse-hcl --compare-to infracost-previous-run.json` to generate a diff")
		}
	}

	return nil
}
