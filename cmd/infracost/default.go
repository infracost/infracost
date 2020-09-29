package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/infracost/infracost/internal/providers/terraform"
	"github.com/infracost/infracost/pkg/config"
	"github.com/infracost/infracost/pkg/output"
	"github.com/infracost/infracost/pkg/prices"
	"github.com/infracost/infracost/pkg/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func defaultCmd() *cli.Command {
	return &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      "tfjson",
				Usage:     "Path to Terraform Plan JSON file",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:      "tfplan",
				Usage:     "Path to Terraform Plan file. Requires 'tfdir' to be set",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:        "tfdir",
				Usage:       "Path to the Terraform project directory",
				TakesFile:   true,
				Value:       getcwd(),
				DefaultText: "current working directory",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output format (json, table)",
				Value:   "table",
			},
			&cli.BoolFlag{
				Name:  "show-skipped",
				Usage: "Prints the list of free and unsupported resources",
				Value: false,
			},
		},
		Action: func(c *cli.Context) error {
			if !checkApiKey() {
				return fmt.Errorf("")
			}

			provider := terraform.New()
			if err := provider.ProcessArgs(c); err != nil {
				return customError(c, err.Error(), false)
			}

			calcSpinner = spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))

			if !config.Config.IsLogging() {
				calcSpinner.Suffix = " Calculating costs…"
				if !c.Bool("no-color") {
					_ = calcSpinner.Color("fgHiGreen", "bold")
				}
				calcSpinner.Start()
			} else {
				log.Info("Calculating costs…")
			}

			resources, err := provider.LoadResources()
			if err != nil {
				return errors.Wrap(err, "error loading resources")
			}

			if err := prices.PopulatePrices(resources); err != nil {
				if e := unwrapped(err); errors.Is(e, prices.InvalidAPIKeyError) {
					calcSpinner.Stop()
					ret := customError(c, e.Error(), false)
					fmt.Println("Please check your INFRACOST_API_KEY environment variable. If you continue having issues please email hello@infracost.io")
					return ret
				}
				return errors.Wrap(err, "error retrieving prices")
			}

			schema.CalculateCosts(resources)

			schema.SortResources(resources)

			var out []byte
			switch strings.ToLower(c.String("output")) {
			case "json":
				out, err = output.ToJSON(resources, c)
			default:
				out, err = output.ToTable(resources, c)
			}

			calcSpinner.Stop()

			if err != nil {
				return errors.Wrap(err, "output error")
			}

			fmt.Println(string(out))

			return nil
		},
	}
}

func getcwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Warn(err)
		cwd = ""
	}

	return cwd
}

func unwrapped(err error) error {
	e := err
	for errors.Unwrap(e) != nil {
		e = errors.Unwrap(e)
	}
	return e
}
