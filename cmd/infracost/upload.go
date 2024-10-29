package main

import (
	"fmt"
	"strings"

	jsoniter "github.com/json-iterator/go"

	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/ui"
)

var (
	validUploadOutputFormats = []string{
		"json",
	}
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
		RunE: checkAPIKeyIsValid(ctx, func(cmd *cobra.Command, args []string) error {
			var err error

			format, _ := cmd.Flags().GetString("format")
			format = strings.ToLower(format)
			if format != "" && !contains(validUploadOutputFormats, format) {
				ui.PrintUsage(cmd)
				return fmt.Errorf("--format only supports %s", strings.Join(validOutputFormats, ", "))
			}

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

			if format == "json" {
				b, err := jsoniter.MarshalIndent(result, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal result: %w", err)
				}
				cmd.Print(string(b))
			} else if result.ShareURL != "" {
				cmd.Println("Share this cost estimate: ", ui.LinkString(result.ShareURL))
			}

			pricingClient := apiclient.GetPricingAPIClient(ctx)
			err = pricingClient.AddEvent("infracost-upload", ctx.EventEnv())
			if err != nil {
				logging.Logger.Warn().Err(err).Msg("could not report `infracost-upload` event")
			}

			return nil
		}),
	}

	cmd.Flags().String("path", "p", "Path to Infracost JSON file.")
	cmd.Flags().String("format", "", "Output format: json")

	_ = cmd.MarkFlagRequired("path")
	_ = cmd.MarkFlagFilename("path", "json")
	return cmd
}
