package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/ui"
)

var (
	validOutputFormats = []string{
		"table",
		"diff",
		"json",
		"html",
		"github-comment",
		"gitlab-comment",
		"azure-repos-comment",
		"bitbucket-comment",
		"bitbucket-comment-summary",
		"slack-message",
	}

	validCompareToFormats = map[string]bool{
		"diff":                      true,
		"json":                      true,
		"github-comment":            true,
		"gitlab-comment":            true,
		"azure-repos-comment":       true,
		"bitbucket-comment":         true,
		"bitbucket-comment-summary": true,
		"slack-message":             true,
	}
)

func outputCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "output",
		Short: "Combine and output Infracost JSON files in different formats",
		Long:  "Combine and output Infracost JSON files in different formats",
		Example: `  Show a breakdown from multiple Infracost JSON files:

      infracost output --path out1.json --path out2.json --path out3.json

  Create HTML report from multiple Infracost JSON files:

      infracost output --format html --path "out*.json" --out-file output.html # glob needs quotes

  Merge multiple Infracost JSON files:

      infracost output --format json --path "out*.json" # glob needs quotes

  Create markdown report to post in a GitHub comment:

      infracost output --format github-comment --path "out*.json" # glob needs quotes

  Create markdown report to post in a GitLab comment:

      infracost output --format gitlab-comment --path "out*.json" # glob needs quotes

  Create markdown report to post in a Azure DevOps Repos comment:

      infracost output --format azure-repos-comment --path "out*.json" # glob needs quotes

  Create markdown report to post in a Bitbucket comment:

      infracost output --format bitbucket-comment --path "out*.json" # glob needs quotes`,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			format, _ := cmd.Flags().GetString("format")
			format = strings.ToLower(format)
			ctx.ContextValues.SetValue("outputFormat", format)

			if format != "" && !contains(validOutputFormats, format) {
				ui.PrintUsage(cmd)
				return fmt.Errorf("--format only supports %s", strings.Join(validOutputFormats, ", "))
			}

			paths, _ := cmd.Flags().GetStringArray("path")

			inputs, err := output.LoadPaths(paths)
			if err != nil {
				return err
			}

			combined, err := output.Combine(inputs)
			if errors.As(err, &clierror.WarningError{}) {
				if format == "json" {
					logging.Logger.Warn().Msg(err.Error())
				}
			} else if err != nil {
				return err
			}
			combined.IsCIRun = ctx.IsCIRun()
			combined.Metadata.InfracostCommand = "output"

			includeAllFields := "all"
			validFields := []string{"price", "monthlyQuantity", "unit", "hourlyCost", "monthlyCost"}

			fields := []string{"monthlyQuantity", "unit", "monthlyCost"}
			if cmd.Flags().Changed("fields") {
				fields, _ = cmd.Flags().GetStringSlice("fields")
				if len(fields) == 0 {
					logging.Logger.Warn().Msgf("fields is empty, using defaults: %s", cmd.Flag("fields").DefValue)
				} else if len(fields) == 1 && fields[0] == includeAllFields {
					fields = validFields
				} else {
					vf := []string{}
					for _, f := range fields {
						if !contains(validFields, f) {
							logging.Logger.Warn().Msgf("Invalid field '%s' specified, valid fields are: %s or '%s' to include all fields", f, validFields, includeAllFields)
						} else {
							vf = append(vf, f)
						}
					}
					fields = vf
				}
			}

			opts := output.Options{
				DashboardEndpoint: ctx.Config.DashboardEndpoint,
				NoColor:           ctx.Config.NoColor,
				Fields:            fields,
				CurrencyFormat:    ctx.Config.CurrencyFormat,
			}
			opts.ShowSkipped, _ = cmd.Flags().GetBool("show-skipped")
			opts.ShowAllProjects, _ = cmd.Flags().GetBool("show-all-projects")

			validFieldsFormats := []string{"table", "html"}

			if cmd.Flags().Changed("fields") && !contains(validFieldsFormats, format) {
				logging.Logger.Warn().Msg("fields is only supported for table and html output formats")
			}

			if ctx.IsCloudUploadExplicitlyEnabled() {
				if ctx.Config.IsSelfHosted() {
					logging.Logger.Warn().Msg("Infracost Cloud is part of Infracost's hosted services. Contact hello@infracost.io for help.")
				} else {
					result := shareCombinedRun(ctx, combined, inputs)
					combined.RunID, combined.ShareURL, combined.CloudURL = result.RunID, result.ShareURL, result.CloudURL
				}
			}

			b, err := output.FormatOutput(format, combined, opts)
			if err != nil {
				return err
			}

			pricingClient := apiclient.GetPricingAPIClient(ctx)
			err = pricingClient.AddEvent("infracost-output", ctx.EventEnv())
			if err != nil {
				logging.Logger.Error().Msgf("Error reporting event: %s", err)
			}

			if outFile, _ := cmd.Flags().GetString("out-file"); outFile != "" {
				err = saveOutFile(ctx, cmd, outFile, b)
				if err != nil {
					return err
				}
			} else {
				cmd.Println(string(b))
			}

			return nil
		},
	}

	cmd.Flags().StringArrayP("path", "p", []string{}, "Path to Infracost JSON files, glob patterns need quotes")
	cmd.Flags().StringP("out-file", "o", "", "Save output to a file, helpful with format flag")

	cmd.Flags().String("format", "table", "Output format: json, diff, table, html, github-comment, gitlab-comment, azure-repos-comment, bitbucket-comment, bitbucket-comment-summary, slack-message")
	cmd.Flags().Bool("show-all-projects", false, "Show all projects in the table of the comment output")
	cmd.Flags().Bool("show-skipped", false, "List unsupported resources")
	cmd.Flags().StringSlice("fields", []string{"monthlyQuantity", "unit", "monthlyCost"}, "Comma separated list of output fields: all,price,monthlyQuantity,unit,hourlyCost,monthlyCost.\nSupported by table and html output formats")

	_ = cmd.MarkFlagRequired("path")
	_ = cmd.MarkFlagFilename("path", "json")

	_ = cmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validOutputFormats, cobra.ShellCompDirectiveDefault
	})

	return cmd
}

func shareCombinedRun(ctx *config.RunContext, combined output.Root, inputs []output.ReportInput) apiclient.AddRunResponse {
	combinedRunIds := []string{}
	for _, input := range inputs {
		if id := input.Root.RunID; id != "" {
			combinedRunIds = append(combinedRunIds, id)
		}
	}
	ctx.ContextValues.SetValue("runIds", combinedRunIds)

	dashboardClient := apiclient.NewDashboardAPIClient(ctx)
	result, err := dashboardClient.AddRun(ctx, combined)
	if err != nil {
		logging.Logger.Err(err).Msg("Failed to upload to Infracost Cloud")
	}

	return result
}

func contains(arr []string, e string) bool {
	return slices.Contains(arr, e)
}
