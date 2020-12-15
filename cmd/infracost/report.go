package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/infracost/infracost/internal/output"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func reportCmd() *cli.Command {
	return &cli.Command{
		Name:  "report",
		Usage: "Create a report from multiple infracost JSON files",
		UsageText: `infracost report [command options] [JSON paths...]

EXAMPLES:
	# Create report from multiple infracost JSON files
	infracost report out1.json out2.json out3.json

	# Create HTML report from multiple infracost JSON files
	infracost report --output html out*.json > report.html

	# Merge multiple infracost JSON files
	infracost report --output json out*.json`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output format: json, table, html",
				Value:   "table",
			},
			&cli.BoolFlag{
				Name:  "show-skipped",
				Usage: "Show unsupported resources, some of which might be free. Only for table and HTML output",
				Value: false,
			},
		},
		Action: func(c *cli.Context) error {
			if c.String("output") == "json" && c.Bool("show-skipped") {
				msg := color.YellowString("The --show-skipped option is not needed with JSON output as that always includes them\n")
				fmt.Fprint(os.Stderr, msg)
			}

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
						"filename": path.Base(f),
					},
					Root: j,
				})
			}

			opts := output.Options{GroupKey: "filename", GroupLabel: "File"}
			combined := output.Combine(inputs, opts)

			var (
				b   []byte
				err error
			)
			switch strings.ToLower(c.String("output")) {
			case "json":
				b, err = output.ToJSON(combined)
			case "html":
				b, err = output.ToHTML(combined, opts, c)
			default:
				b, err = output.ToTable(combined, c)
			}
			if err != nil {
				return err
			}

			fmt.Println(string(b))

			return nil
		},
	}
}
