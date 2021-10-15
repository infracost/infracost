package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/ui"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

var minOutputVersion = "0.2"
var maxOutputVersion = "0.2"

func outputCmd(ctx *config.RunContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "output",
		Short: "Combine and output Infracost JSON files in different formats",
		Long:  "Combine and output Infracost JSON files in different formats",
		Example: `  Show a breakdown from multiple Infracost JSON files:

      infracost output --path out1.json --path out2.json --path out3.json

  Create HTML report from multiple Infracost JSON files:

      infracost output --format html --path "out*.json" > output.html

  Merge multiple Infracost JSON files:

      infracost output --format json --path "out*.json"`,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFiles := []string{}

			paths, _ := cmd.Flags().GetStringArray("path")
			for _, path := range paths {
				expanded, err := homedir.Expand(path)
				if err != nil {
					return errors.Wrap(err, "Failed to expand path")
				}

				matches, _ := filepath.Glob(expanded)
				if len(matches) > 0 {
					inputFiles = append(inputFiles, matches...)
				} else {
					inputFiles = append(inputFiles, path)
				}
			}

			inputs := make([]output.ReportInput, 0, len(inputFiles))
			currency := ""

			for _, f := range inputFiles {
				data, err := os.ReadFile(f)
				if err != nil {
					return errors.Wrap(err, "Error reading JSON file")
				}

				j, err := output.Load(data)
				if err != nil {
					return errors.Wrap(err, "Error parsing JSON file")
				}

				if !checkOutputVersion(j.Version) {
					return fmt.Errorf("Invalid Infracost JSON file version. Supported versions are %s ≤ x ≤ %s", minOutputVersion, maxOutputVersion)
				}

				currency, err = checkCurrency(currency, j.Currency)
				if err != nil {
					return err
				}

				inputs = append(inputs, output.ReportInput{
					Metadata: map[string]string{
						"filename": f,
					},
					Root: j,
				})
			}

			format, _ := cmd.Flags().GetString("format")
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
				GroupKey:         "filename",
				GroupLabel:       "File",
				Fields:           fields,
			}
			opts.ShowSkipped, _ = cmd.Flags().GetBool("show-skipped")

			combined := output.Combine(currency, inputs, opts)

			var (
				b   []byte
				err error
			)

			validFieldsFormats := []string{"table", "html"}

			if cmd.Flags().Changed("fields") && !contains(validFieldsFormats, format) {
				ui.PrintWarning(cmd.ErrOrStderr(), "fields is only supported for table and html output formats")
			}
			switch strings.ToLower(format) {
			case "json":
				b, err = output.ToJSON(combined, opts)
			case "html":
				b, err = output.ToHTML(combined, opts)
			case "diff":
				b, err = output.ToDiff(combined, opts)
			default:
				b, err = output.ToTable(combined, opts)
			}
			if err != nil {
				return err
			}

			cmd.Println(string(b))

			return nil
		},
	}

	cmd.Flags().StringArrayP("path", "p", []string{}, "Path to Infracost JSON files")
	_ = cmd.MarkFlagRequired("path")
	_ = cmd.MarkFlagFilename("path", "json")

	cmd.Flags().String("format", "table", "Output format: json, diff, table, html")
	cmd.Flags().Bool("show-skipped", false, "Show unsupported resources, some of which might be free")
	cmd.Flags().StringSlice("fields", []string{"monthlyQuantity", "unit", "monthlyCost"}, "Comma separated list of output fields: all,price,monthlyQuantity,unit,hourlyCost,monthlyCost.\nSupported by table and html output formats")

	_ = cmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"table", "json", "html"}, cobra.ShellCompDirectiveDefault
	})

	return cmd
}

func checkCurrency(inputCurrency, fileCurrency string) (string, error) {
	if fileCurrency == "" {
		fileCurrency = "USD" // default to USD
	}

	if inputCurrency == "" {
		// this must be the first file, save the input currency
		inputCurrency = fileCurrency
	}

	if inputCurrency != fileCurrency {
		return "", fmt.Errorf("Invalid Infracost JSON file currency mismatch.  Can't combine %s and %s", inputCurrency, fileCurrency)
	}

	return inputCurrency, nil
}

func checkOutputVersion(v string) bool {
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	return semver.Compare(v, "v"+minOutputVersion) >= 0 && semver.Compare(v, "v"+maxOutputVersion) <= 0
}

func contains(arr []string, e string) bool {
	for _, a := range arr {
		if a == e {
			return true
		}
	}
	return false
}
