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
		},
		Action: func(c *cli.Context) error {
			resources, err := terraform.ParsePlanFile(c.String("plan"))
			if err != nil {
				return err
			}

			resourceCostBreakdowns, err := base.GetCostBreakdowns(resources)
			if err != nil {
				return err
			}

			jsonBytes, err := outputs.ToJSON(resourceCostBreakdowns)
			if err != nil {
				return err
			}
			fmt.Println(string(jsonBytes))

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
