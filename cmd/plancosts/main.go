package main

import (
	"fmt"
	"os"

	"plancosts/pkg/base"
	"plancosts/pkg/config"
	"plancosts/pkg/output"
	"plancosts/pkg/parsers/terraform"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "plancosts",
		Usage: "Generate cost reports from Terraform plans",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      "tfplan-json",
				Usage:     "Path to Terraform Plan JSON file",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:      "tfplan",
				Usage:     "Path to Terraform Plan file. Requires tfpath to also be set",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:  "tfpath",
				Usage: "Path to the Terraform project",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Verbosity",
				Value:   false,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output (json|table)",
				Value:   "table",
			},
		},
		Action: func(c *cli.Context) error {
			var planJSON []byte

			logLevel := log.InfoLevel
			if c.Bool("verbose") {
				logLevel = log.DebugLevel
			}
			log.SetLevel(logLevel)

			if c.String("tfplan-json") != "" && c.String("tfplan") != "" {
				color.HiRed("Please only provide one of either a Terraform Plan JSON file (tfplan-json) or a Terraform Plan file (tfplan)")
				_ = cli.ShowAppHelp(c)
				os.Exit(1)
			}

			if c.String("tfplan") != "" && c.String("tfpath") == "" {
				color.HiRed("Please provide a path to the Terrafrom project (tfpath) if providing a Terraform Plan file (tfplan)\n\n")
				_ = cli.ShowAppHelp(c)
				os.Exit(1)
			}

			if c.String("tfplan-json") == "" && c.String("tfpath") == "" {
				color.HiRed("Please provide either the path to the Terrafrom project (tfpath) or a Terraform Plan JSON file (tfplan-json)")
				_ = cli.ShowAppHelp(c)
				os.Exit(1)
			}

			var err error
			if c.String("tfplan-json") != "" {
				planJSON, err = terraform.LoadPlanJSON(c.String("tfplan-json"))
				if err != nil {
					return err
				}
			} else {
				planFile := c.String("tfplan")
				planJSON, err = terraform.GeneratePlanJSON(c.String("tfpath"), planFile)
				if err != nil {
					return err
				}
			}

			resources, err := terraform.ParsePlanJSON(planJSON)
			if err != nil {
				return err
			}

			q := base.NewGraphQLQueryRunner(config.Config.PriceListApiEndpoint)
			resourceCostBreakdowns, err := base.GenerateCostBreakdowns(q, resources)
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
		os.Exit(1)
	}
}
