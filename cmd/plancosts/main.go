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
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
	})

	app := &cli.App{
		Name:                 "plancosts",
		Usage:                "Generate cost reports from Terraform plans",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      "tfplan-json",
				Usage:     "Path to Terraform Plan JSON file",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:      "tfplan",
				Usage:     "Path to Terraform Plan file. Requires tfdir to also be set",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:      "tfdir",
				Usage:     "Path to the Terraform project directory",
				TakesFile: true,
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
			&cli.StringFlag{
				Name:  "api-url",
				Usage: "Price List API URL",
			},
		},
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			log.Error(err)
			_ = cli.ShowAppHelp(c)
			os.Exit(1)
			return nil
		},
		Action: func(c *cli.Context) error {
			var planJSON []byte

			logLevel := log.InfoLevel
			if c.Bool("verbose") {
				logLevel = log.DebugLevel
			}
			log.SetLevel(logLevel)

			if c.String("api-url") != "" {
				config.Config.ApiUrl = c.String("api-url")
			}

			if c.String("tfplan-json") != "" && c.String("tfplan") != "" {
				color.HiRed("Please only provide one of either a Terraform Plan JSON file (tfplan-json) or a Terraform Plan file (tfplan)")
				_ = cli.ShowAppHelp(c)
				os.Exit(1)
			}

			if c.String("tfplan") != "" && c.String("tfdir") == "" {
				color.HiRed("Please provide a path to the Terrafrom project (tfdir) if providing a Terraform Plan file (tfplan)\n\n")
				_ = cli.ShowAppHelp(c)
				os.Exit(1)
			}

			if c.String("tfplan-json") == "" && c.String("tfdir") == "" {
				color.HiRed("Please provide either the path to the Terrafrom project (tfdir) or a Terraform Plan JSON file (tfplan-json)")
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
				planJSON, err = terraform.GeneratePlanJSON(c.String("tfdir"), planFile)
				if err != nil {
					return err
				}
			}

			resources, err := terraform.ParsePlanJSON(planJSON)
			if err != nil {
				return err
			}

			q := base.NewGraphQLQueryRunner(fmt.Sprintf("%s/graphql", config.Config.ApiUrl))
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
		log.Error(err)
		os.Exit(1)
	}
}
