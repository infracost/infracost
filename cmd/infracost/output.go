package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/fatih/color"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func outputCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "output",
		Short: "Output Infracost JSON files in different formats",
		Long: `Output Infracost JSON files in different formats

Show a breakdown from multiple Infracost JSON files:

  infracost output out1.json out2.json out3.json

Create HTML report from multiple Infracost JSON files:

  infracost output --format html out*.json > output.html

Merge multiple Infracost JSON files:

  infracost output --format json out*.json`,
		PreRun: func(cmd *cobra.Command, args []string) {
			handleDeprecatedEnvVars(deprecatedEnvVarMapping)
			handleDeprecatedFlags(cmd, deprecatedFlagsMapping)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			files := args

			inputs := make([]output.ReportInput, 0, len(files))
			for _, f := range files {
				data, err := ioutil.ReadFile(f)
				if err != nil {
					return errors.Wrap(err, "Error reading JSON file")
				}

				j, err := output.Load(data)
				if err != nil {
					return errors.Wrap(err, "Error parsing JSON file")
				}

				inputs = append(inputs, output.ReportInput{
					Metadata: map[string]string{
						"filename": f,
					},
					Root: j,
				})
			}

			format, _ := cmd.Flags().GetString("format")

			opts := output.Options{
				NoColor:    cfg.NoColor,
				GroupKey:   "filename",
				GroupLabel: "File",
			}
			opts.ShowSkipped, _ = cmd.Flags().GetBool("show-skipped")

			combined := output.Combine(inputs, opts)

			var (
				b   []byte
				err error
			)
			switch strings.ToLower(format) {
			case "json":
				b, err = output.ToJSON(combined, opts)
			case "html":
				b, err = output.ToHTML(combined, opts)
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

	cmd.Flags().String("format", "table", "Output format: json, diff, table, html")
	cmd.Flags().Bool("show-skipped", false, "Show unsupported resources, some of which might be free. Ignored for JSON outputs")

	return cmd
}

func reportCmd(cfg *config.Config) *cobra.Command {
	deprecationMsg := "This command is deprecated and will be removed in v0.8.0. Please use `infracost output`."

	cmd := outputCmd(cfg)
	cmd.Use = "report"
	cmd.Hidden = true
	cmd.Long = color.YellowString(deprecationMsg)

	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		usageWarning(deprecationMsg)
	}

	// Add deprecated flag
	cmd.Flags().StringP("output", "o", "table", "Output format: json, table, html")
	_ = cmd.Flags().MarkHidden("output")

	return cmd
}
