package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/ui"
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

      infracost output --format html --path out*.json > output.html

  Merge multiple Infracost JSON files:

      infracost output --format json --path out*.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFiles := []string{}

			if !cmd.Flags().Changed("path") {
				m := fmt.Sprintf("No path specified\n\nUse the %s flag to specify the path to an Infracost JSON file.", ui.PrimaryString("--path"))
				ui.PrintUsageErrorAndExit(cmd, m)
			}

			paths, _ := cmd.Flags().GetStringArray("path")
			for _, path := range paths {
				matches, _ := filepath.Glob(path)
				inputFiles = append(inputFiles, matches...)
			}

			inputs := make([]output.ReportInput, 0, len(inputFiles))
			for _, f := range inputFiles {
				data, err := ioutil.ReadFile(f)
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

				inputs = append(inputs, output.ReportInput{
					Metadata: map[string]string{
						"filename": f,
					},
					Root: j,
				})
			}

			format, _ := cmd.Flags().GetString("format")

			validFields := []string{"price", "monthlyQuantity", "unit", "hourlyCost", "monthlyCost"}

			fields := []string{"monthlyQuantity", "unit", "monthlyCost"}
			if cmd.Flags().Changed("fields") {
				if c, _ := cmd.Flags().GetStringSlice("fields"); len(c) == 0 {
					ui.PrintWarningf("fields is empty, using defaults: %s", cmd.Flag("fields").DefValue)
				} else {
					fields, _ = cmd.Flags().GetStringSlice("fields")
					vf := []string{}
					for _, f := range fields {
						if !contains(validFields, f) {
							ui.PrintWarningf("Invalid field '%s' specified, valid fields are: %s", f, validFields)
						} else {
							vf = append(vf, f)
						}
					}
					fields = vf
				}
			}

			opts := output.Options{
				NoColor:    ctx.Config.NoColor,
				GroupKey:   "filename",
				GroupLabel: "File",
				Fields:     fields,
			}
			opts.ShowSkipped, _ = cmd.Flags().GetBool("show-skipped")

			combined := output.Combine(inputs, opts)

			var (
				b   []byte
				err error
			)

			validFieldsFormats := []string{"table", "html"}

			if cmd.Flags().Changed("fields") && !contains(validFieldsFormats, format) {
				ui.PrintWarning("fields is only supported for table and html output formats")
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

			fmt.Println(string(b))

			return nil
		},
	}

	cmd.Flags().StringArrayP("path", "p", []string{}, "Path to Infracost JSON files")

	cmd.Flags().String("format", "table", "Output format: json, diff, table, html")
	cmd.Flags().Bool("show-skipped", false, "Show unsupported resources, some of which might be free")
	cmd.Flags().StringSlice("fields", []string{"monthlyQuantity", "unit", "monthlyCost"}, "Comma separated list of output fields: price,monthlyQuantity,unit,hourlyCost,monthlyCost.\nSupported by table and html output formats")

	return cmd
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
