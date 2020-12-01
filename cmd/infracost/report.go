package main

import (
	"fmt"
	"io/ioutil"
	"strings"

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
	# Run with multiple infracost JSON files
	infracost report out1.json out2.json

	# Merge multiple JSON files into a single JSON output
	infracost report --output json out1.json out2.json`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output format (json, table, html)",
				Value:   "table",
			},
		},
		Action: func(c *cli.Context) error {
			files := make([]string, 0)

			for i := 0; i < c.Args().Len(); i++ {
				files = append(files, c.Args().Get(i))
			}

			jsons := make([]output.Root, len(files))
			for _, f := range files {
				data, err := ioutil.ReadFile(f)
				if err != nil {
					return errors.Wrap(err, "Error reading JSON file")
				}

				j, err := output.Load(data)
				if err != nil {
					return errors.Wrap(err, "Error parsing JSON file")
				}

				jsons = append(jsons, j)
			}

			combined := output.Combine(jsons...)

			var (
				b   []byte
				err error
			)
			switch strings.ToLower(c.String("output")) {
			case "json":
				b, err = output.ToJSON(combined)
			case "html":
				b, err = output.ToHTML(combined)
			default:
				b, err = output.ToTable(combined)
			}
			if err != nil {
				return err
			}

			fmt.Println(string(b))

			return nil
		},
	}
}
