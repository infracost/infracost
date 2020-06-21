package main

import (
	"fmt"
	"log"
	"os"

	"plancosts/pkg/base"
	"plancosts/pkg/output"
	"plancosts/pkg/parsers/terraform"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "plancosts",
		Usage: "Generate cost reports from Terraform plans",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "plan, p",
				Usage:    "Path to Terraform Plan JSON",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "output, o",
				Usage: "Output (json|table)",
				Value: "table",
			},
		},
		Action: func(c *cli.Context) error {
			resources, err := terraform.ParsePlanFile(c.String("plan"))
			if err != nil {
				return err
			}

			resourceCostBreakdowns, err := base.GenerateCostBreakdowns(resources)
			if err != nil {
				return err
			}

			var out []byte
			switch c.String("output") {
			case "table":
				out, err = output.ToTable(resourceCostBreakdowns)
			case "json":
				out, err = output.ToJSON(resourceCostBreakdowns)
			default:
				err = cli.ShowAppHelp(c)
			}
			if err != nil {
				return err
			}

			fmt.Println(string(out))

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
