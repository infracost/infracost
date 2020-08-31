package main

import (
	"fmt"
	"os"

	"infracost/internal/providers/terraform"
	"infracost/pkg/config"
	"infracost/pkg/output"
	"infracost/pkg/prices"
	"infracost/pkg/schema"

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

			provider := terraform.Provider()
			err := provider.ProcessArgs(c)
			if err != nil {
				color.HiRed(err.Error())
				_ = cli.ShowAppHelp(c)
				os.Exit(1)
			}

			resources, err := provider.LoadResources()
			if err != nil {
				return err
			}
			err = prices.PopulatePrices(resources)
			if err != nil {
				return err
			}
			schema.CalculateCosts(resources)

			var out []byte
			switch c.String("output") {
			case "table":
				out, err = output.ToTable(resources)
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
