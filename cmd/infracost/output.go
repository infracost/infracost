package main

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/ui"
)

var validOutputFormats = []string{
	"table",
	"diff",
	"json",
	"html",
	"github-comment",
	"gitlab-comment",
	"azure-repos-comment",
	"bitbucket-comment",
	"slack-message",
}

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
			ctx.SetContextValue("outputFormat", format)

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
			if err != nil {
				return err
			}
			combined.IsCIRun = ctx.IsCIRun()

			includeAllFields := "all"
			validFields := []string{"price", "monthlyQuantity", "unit", "hourlyCost", "monthlyCost"}

			fields := []string{"monthlyQuantity", "unit", "monthlyCost"}
			if cmd.Flags().Changed("fields") {
				fields, _ = cmd.Flags().GetStringSlice("fields")
				if len(fields) == 0 {
					ui.PrintWarningf(cmd.ErrOrStderr(), "fields is empty, using defaults: %s", cmd.Flag("fields").DefValue)
				} else if len(fields) == 1 && fields[0] == includeAllFields {
					fields = validFields
				} else {
					vf := []string{}
					for _, f := range fields {
						if !contains(validFields, f) {
							ui.PrintWarningf(cmd.ErrOrStderr(), "Invalid field '%s' specified, valid fields are: %s or '%s' to include all fields", f, validFields, includeAllFields)
						} else {
							vf = append(vf, f)
						}
					}
					fields = vf
				}
			}

			opts := output.Options{
				DashboardEnabled: ctx.Config.EnableDashboard,
				NoColor:          ctx.Config.NoColor,
				Fields:           fields,
			}
			opts.ShowSkipped, _ = cmd.Flags().GetBool("show-skipped")

			validFieldsFormats := []string{"table", "html"}

			if cmd.Flags().Changed("fields") && !contains(validFieldsFormats, format) {
				ui.PrintWarning(cmd.ErrOrStderr(), "fields is only supported for table and html output formats")
			}

			if ctx.Config.EnableDashboard {
				if ctx.Config.IsSelfHosted() {
					ui.PrintWarning(cmd.ErrOrStderr(), "The dashboard is part of Infracost's hosted services. Contact hello@infracost.io for help.")
				}

				combined.RunID, combined.ShareURL = shareCombinedRun(ctx, combined, inputs)
			}

			var b []byte

			switch strings.ToLower(format) {
			case "json":
				b, err = output.ToJSON(combined, opts)
			case "html":
				b, err = output.ToHTML(combined, opts)
			case "diff":
				b, err = output.ToDiff(combined, opts)
			case "github-comment", "gitlab-comment", "azure-repos-comment":
				b, err = output.ToMarkdown(combined, opts, output.MarkdownOptions{})
			case "bitbucket-comment":
				b, err = output.ToMarkdown(combined, opts, output.MarkdownOptions{BasicSyntax: true})
			case "slack-message":
				b, err = output.ToSlackMessage(combined, opts)
			default:
				b, err = output.ToTable(combined, opts)
			}
			if err != nil {
				return err
			}

			pricingClient := apiclient.NewPricingAPIClient(ctx)
			err = pricingClient.AddEvent("infracost-output", ctx.EventEnv())
			if err != nil {
				log.Errorf("Error reporting event: %s", err)
			}

			if outFile, _ := cmd.Flags().GetString("out-file"); outFile != "" {
				err = saveOutFile(cmd, outFile, b)
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

	cmd.Flags().String("format", "table", "Output format: json, diff, table, html, github-comment, gitlab-comment, azure-repos-comment, bitbucket-comment, slack-message")
	cmd.Flags().Bool("show-skipped", false, "Show unsupported resources")
	cmd.Flags().StringSlice("fields", []string{"monthlyQuantity", "unit", "monthlyCost"}, "Comma separated list of output fields: all,price,monthlyQuantity,unit,hourlyCost,monthlyCost.\nSupported by table and html output formats")

	_ = cmd.MarkFlagRequired("path")
	_ = cmd.MarkFlagFilename("path", "json")

	_ = cmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validOutputFormats, cobra.ShellCompDirectiveDefault
	})

	return cmd
}

func shareCombinedRun(ctx *config.RunContext, combined output.Root, inputs []output.ReportInput) (string, string) {
	if len(inputs) == 1 && inputs[0].Root.RunID != "" {
		result := inputs[0].Root
		return result.RunID, result.ShareURL
	}

	projectContexts := make([]*config.ProjectContext, len(combined.Projects))
	for i := range combined.Projects {
		projectContexts[i] = config.EmptyProjectContext()
	}

	combinedRunIds := []string{}
	for _, input := range inputs {
		if id := input.Root.RunID; id != "" {
			combinedRunIds = append(combinedRunIds, id)
		}
	}
	ctx.SetContextValue("runIds", combinedRunIds)

	dashboardClient := apiclient.NewDashboardAPIClient(ctx)
	result, err := dashboardClient.AddRun(ctx, projectContexts, combined)
	if err != nil {
		log.Errorf("Error reporting run: %s", err)
	}

	return result.RunID, result.ShareURL
}

func contains(arr []string, e string) bool {
	for _, a := range arr {
		if a == e {
			return true
		}
	}
	return false
}
