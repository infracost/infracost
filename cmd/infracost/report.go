package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func reportCmd(cfg *config.Config) *cli.Command {
	deprecatedFlags := []cli.Flag{
		&cli.StringFlag{
			Name:   "output",
			Usage:  "Output format: json, table, html",
			Value:  "table",
			Hidden: true,
		},
	}

	flags := append(deprecatedFlags,
		&cli.StringFlag{
			Name:  "format",
			Usage: "Output format: json, table, html",
			Value: "table",
		},
		&cli.BoolFlag{
			Name:  "show-skipped",
			Usage: "Show unsupported resources, some of which might be free. Only for table and HTML output",
			Value: false,
		},
		&cli.BoolFlag{
			Name:  "no-color",
			Usage: "Turn off colored output",
			Value: false,
		},
	)

	return &cli.Command{
		Name:  "report",
		Usage: "Create a report from multiple Infracost JSON files",
		UsageText: `infracost report [command options] [JSON paths...]

EXAMPLES:
	# Create report from multiple Infracost JSON files
	infracost report out1.json out2.json out3.json

	# Create HTML report from multiple Infracost JSON files
	infracost report --format html out*.json > report.html

	# Merge multiple Infracost JSON files
	infracost report --format json out*.json`,
		Flags: flags,
		Action: func(c *cli.Context) error {
			handleDeprecatedEnvVars(c, deprecatedEnvVarMapping)
			handleDeprecatedFlags(c, deprecatedFlagsMapping)

			files := make([]string, 0)

			for i := 0; i < c.Args().Len(); i++ {
				files = append(files, c.Args().Get(i))
			}

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

			opts := output.Options{
				ShowSkipped: c.Bool("show-skipped"),
				NoColor:     c.Bool("no-color"),
				GroupKey:    "filename",
				GroupLabel:  "File",
			}

			combined := output.Combine(inputs, opts)

			outputCfg := &config.Output{
				Format:      c.String("format"),
				ShowSkipped: c.Bool("show-skipped"),
				NoColor:     c.Bool("no-color"),
			}

			var (
				b   []byte
				err error
			)
			switch strings.ToLower(outputCfg.Format) {
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
}
