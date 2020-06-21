package main

import (
	"fmt"
	"log"
	"os"

	"plancosts/pkg/base"
	"plancosts/pkg/outputs"
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

			var output []byte
			switch c.String("output") {
			case "table":
				output, err = outputs.ToTable(resourceCostBreakdowns)
			case "json":
				output, err = outputs.ToJSON(resourceCostBreakdowns)
			default:
				err = cli.ShowAppHelp(c)
			}
			if err != nil {
				return err
			}

			fmt.Println(string(output))

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
