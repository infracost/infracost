package main

import (
	"fmt"
	"os"

	"infracost/pkg/config"
	"infracost/pkg/costs"
	"infracost/pkg/output"
	"infracost/pkg/terraform"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	formatter := &log.TextFormatter{
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
	}
	log.SetFormatter(formatter)

	app := &cli.App{
		Name:                 "infracost",
		Usage:                "Generate cost reports from Terraform plans",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      "tfjson",
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
			&cli.BoolFlag{
				Name:  "no-color",
				Usage: "Turn off colored output",
				Value: false,
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

			if c.Bool("no-color") {
				config.Config.NoColor = true
				formatter.DisableColors = true
				color.NoColor = true
			}

			logLevel := log.InfoLevel
			if c.Bool("verbose") {
				logLevel = log.DebugLevel
			}
			log.SetLevel(logLevel)

			if c.String("api-url") != "" {
				config.Config.ApiUrl = c.String("api-url")
			}

			if c.String("tfjson") != "" && c.String("tfplan") != "" {
				color.HiRed("Please only provide one of either a Terraform Plan JSON file (tfjson) or a Terraform Plan file (tfplan)")
				_ = cli.ShowAppHelp(c)
				os.Exit(1)
			}

			if c.String("tfplan") != "" && c.String("tfdir") == "" {
				color.HiRed("Please provide a path to the Terrafrom project (tfdir) if providing a Terraform Plan file (tfplan)\n\n")
				_ = cli.ShowAppHelp(c)
				os.Exit(1)
			}

			if c.String("tfjson") == "" && c.String("tfdir") == "" {
				color.HiRed("Please provide either the path to the Terrafrom project (tfdir) or a Terraform Plan JSON file (tfjson)")
				_ = cli.ShowAppHelp(c)
				os.Exit(1)
			}

			var err error
			if c.String("tfjson") != "" {
				planJSON, err = terraform.LoadPlanJSON(c.String("tfjson"))
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

			terraform.ParsePlanJSON(planJSON)
			schemaResources := terraform.ParsePlanJSON(planJSON)

			q := costs.NewGraphQLQueryRunner(fmt.Sprintf("%s/graphql", config.Config.ApiUrl))

			costResources := make([]*costs.Resource, 0, len(schemaResources))
			for _, schemaResource := range schemaResources {
				costResource := costs.NewResource(schemaResource)
				err := costResource.CalculateCosts(q)
				if err != nil {
					return err
				}
				costResources = append(costResources, costResource)
			}

			var out []byte
			switch c.String("output") {
			case "table":
				out, err = output.ToTable(costResources)
			// TODO
			// case "json":
			// 	out, err = output.ToJSON(costResources)
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
