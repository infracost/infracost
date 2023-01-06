package main

import (
	"fmt"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/ui"
	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
)

func uploadCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload an Infracost JSON file to Infracost Cloud",
		Long: `Upload an Infracost JSON file to Infracost Cloud. This is useful if you
do not use 'infracost comment' and instead want to define run metadata,
such as pull request URL or title, and upload the results manually.

See https://infracost.io/docs/features/cli_commands/#upload-runs`,
		Example: `  Upload an Infracost JSON file:
      export INFRACOST_VCS_PULL_REQUEST_URL=http://github.com/myorg...
      export INFRACOST_VCS_PULL_REQUEST_TITLE="My PR title"
      # ... other env vars here

      infracost diff --path plan.json --format json --out-file infracost.json

      infracost upload --path infracost.json`,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if ctx.Config.IsSelfHosted() {
				return fmt.Errorf("Infracost Cloud is part of Infracost's hosted services. Contact hello@infracost.io for help.")
			}

			path, _ := cmd.Flags().GetString("path")

			root, err := output.Load(path)
			if err != nil {
				return fmt.Errorf("could not load input file %s err: %w", path, err)
			}

			dashboardClient := apiclient.NewDashboardAPIClient(ctx)
			result, err := dashboardClient.AddRun(ctx, root)
			if err != nil {
				return fmt.Errorf("failed to upload to Infracost Cloud: %w", err)
			}

			root.RunID, root.ShareURL = result.RunID, result.ShareURL

			if root.ShareURL != "" {
				cmd.Println("Share this cost estimate: ", ui.LinkString(root.ShareURL))
			}

			pricingClient := apiclient.NewPricingAPIClient(ctx)
			err = pricingClient.AddEvent("infracost-upload", ctx.EventEnv())
			if err != nil {
				logging.Logger.WithError(err).Warn("could not report `infracost-upload` event")
			}

			if len(result.GuardrailCheck.BlockingFailures) > 0 {
				return result.GuardrailCheck.BlockingFailures
			}

			return nil
		},
	}

	cmd.Flags().String("path", "p", "Path to Infracost JSON file.")

	_ = cmd.MarkFlagRequired("path")
	_ = cmd.MarkFlagFilename("path", "json")
	return cmd
}
