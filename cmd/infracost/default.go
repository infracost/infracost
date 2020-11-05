package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/prices"
	"github.com/infracost/infracost/internal/providers/terraform"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/spin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func defaultCmd() *cli.Command {
	return &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      "tfjson",
				Usage:     "Path to Terraform plan JSON file",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:      "tfplan",
				Usage:     "Path to Terraform plan file relative to 'tfdir'. Requires 'tfdir' to be set",
				TakesFile: true,
			},
			&cli.BoolFlag{
				Name:  "use-tfstate",
				Usage: "Use Terraform state instead of generating a plan",
				Value: false,
			},
			&cli.StringFlag{
				Name:        "tfdir",
				Usage:       "Path to the Terraform code directory",
				TakesFile:   true,
				Value:       getcwd(),
				DefaultText: "current working directory",
			},
			&cli.StringFlag{
				Name:  "tfflags",
				Usage: "Flags to pass to the 'terraform plan' command",
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
			if err := checkAPIKey(); err != nil {
				return err
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

				red := color.New(color.FgHiRed)
				bold := color.New(color.Bold, color.FgHiWhite)

				if e := unwrapped(err); errors.Is(e, prices.ErrInvalidAPIKey) {
					return errors.New(fmt.Sprintf("%v\n%s %s %s %s %s\n%s",
						e.Error(),
						red.Sprint("Please check your"),
						bold.Sprint(config.ConfigFilePath()),
						red.Sprint("file or"),
						bold.Sprint("INFRACOST_API_KEY"),
						red.Sprint("environment variable."),
						red.Sprint("If you continue having issues please email hello@infracost.io"),
					))
				}

				if e, ok := err.(*prices.PricingAPIError); ok {
					return errors.New(fmt.Sprintf("%v\n%s", e.Error(), "We have been notified of this issue."))
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
