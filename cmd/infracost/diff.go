package main

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/providers"
	"github.com/infracost/infracost/internal/ui"
)

func diffCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Show diff of monthly costs between current and planned state",
		Long:  "Show diff of monthly costs between current and planned state",
		Example: `  Use Terraform directory:

      infracost breakdown --path /code --format json --out-file infracost-base.json
      # Make Terraform code changes
      infracost diff --path /code --compare-to infracost-base.json

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
		DashboardEndpoint: ctx.Config.DashboardEndpoint,
		ShowSkipped:       ctx.Config.ShowSkipped,
		NoColor:           ctx.Config.NoColor,
		Fields:            ctx.Config.Fields,
	})
	if err != nil {
		return err
	}

	pricingClient := apiclient.NewPricingAPIClient(ctx)
	err = pricingClient.AddEvent("infracost-run", ctx.EventEnv())
	if err != nil {
		logging.Logger.WithError(err).Error("could not report infracost-run event")
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

		projectType := providers.DetectProjectType(projectConfig.Path, projectConfig.TerraformForceCLI)
		if (projectType == "terraform_dir" || projectType == "terragrunt_dir") && cfg.CompareTo == "" {
			examplePath := "/code"
			if projectConfig.Path != "" {
				examplePath = projectConfig.Path
			}

			msg := fmt.Sprintf(`To show a diff:
  1. Generate a cost estimate baseline: %s
  2. Make a Terraform code change
  3. Generate a cost estimate diff: %s`,
				fmt.Sprintf("`infracost breakdown --path %s --format json --out-file infracost-base.json`", examplePath),
				fmt.Sprintf("`infracost diff --path %s --compare-to infracost-base.json`", examplePath),
			)
			return errors.New(msg)
		}
	}

	return nil
}
