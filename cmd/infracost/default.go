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
	"github.com/infracost/infracost/internal/usage"
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
				Usage:     "Path to Terraform plan file relative to 'tfdir'",
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
				Usage:   "Output format: json, table, html",
				Value:   "table",
			},
			&cli.BoolFlag{
				Name:  "show-skipped",
				Usage: "Show unsupported resources, some of which might be free. Only for table and HTML output",
				Value: false,
			},
			&cli.StringFlag{
				Name:      "usage-file",
				Usage:     "Path to Infracost usage file that specifies values for usage-based resources",
				TakesFile: true,
			},
		},
		Action: func(c *cli.Context) error {
			if err := checkAPIKey(); err != nil {
				return err
			}

			config.Environment.Flags = c.FlagNames()
			config.Environment.OutputFormat = c.String("output")

			if c.String("output") == "json" && c.Bool("show-skipped") {
				msg := color.YellowString("The --show-skipped option is not needed with JSON output as that always includes them\n")
				fmt.Fprint(os.Stderr, msg)
			}

			provider := terraform.New()
			if err := provider.ProcessArgs(c); err != nil {
				usageError(c, err.Error())
			}

			u, err := usage.LoadFromFile(c.String("usage-file"))
			if err != nil {
				return err
			}
			if len(u) > 0 {
				config.Environment.HasUsageFile = true
			}

			resources, err := provider.LoadResources(u)
			if err != nil {
				return err
			}

			spinner = spin.NewSpinner("Calculating cost estimate")

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

			opts := output.Options{}
			r := output.ToOutputFormat(resources)
			var (
				b   []byte
				out string
			)
			switch strings.ToLower(c.String("output")) {
			case "json":
				b, err = output.ToJSON(r)
				out = string(b)
			case "html":
				b, err = output.ToHTML(r, opts, c)
				out = string(b)
			default:
				b, err = output.ToTable(r, c)
				out = fmt.Sprintf("\n%s", string(b))
			}

			if err != nil {
				spinner.Fail()
				return errors.Wrap(err, "Error generating output")
			}

			spinner.Success()

			fmt.Printf("%s\n", out)

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
