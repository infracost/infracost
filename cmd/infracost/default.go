package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/infracost/infracost/internal/providers/terraform"
	"github.com/infracost/infracost/internal/spin"
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
				Name:  "tfflags",
				Usage: "Arguments to pass to the 'terraform plan' command",
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
				os.Exit(1)
			}

			provider := terraform.New()
			if err := provider.ProcessArgs(c); err != nil {
				usageError(c, err.Error())
			}

			resources, err := provider.LoadResources()
			if err != nil {
				return err
			}

			spinner = spin.NewSpinner("Calculating costs")

			if err := prices.PopulatePrices(resources); err != nil {
				spinner.Fail()
				if e := unwrapped(err); errors.Is(e, prices.InvalidAPIKeyError) {
					red := color.New(color.FgHiRed)
					bold := color.New(color.Bold, color.FgHiWhite)
					fmt.Fprintln(os.Stderr, red.Sprint(e.Error()))
					fmt.Fprintln(os.Stderr, red.Sprint("Please check your"), bold.Sprint("INFRACOST_API_KEY"), red.Sprint("environment variable. If you continue having issues please email hello@infracost.io"))
					os.Exit(1)
				}
				return err
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

			if err != nil {
				spinner.Fail()
				return errors.Wrap(err, "Error generating output")
			}

			spinner.Success()

			fmt.Printf("\n%s\n", string(out))

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
