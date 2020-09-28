package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/infracost/infracost/pkg/config"
	"github.com/infracost/infracost/pkg/output"
	"github.com/infracost/infracost/pkg/prices"
	"github.com/infracost/infracost/pkg/schema"
	"github.com/infracost/infracost/pkg/version"
	"github.com/pkg/errors"

	"github.com/infracost/infracost/internal/providers/terraform"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func customError(c *cli.Context, msg string) error {
	color.HiRed(fmt.Sprintf("%v\n", msg))
	_ = cli.ShowAppHelp(c)

	return fmt.Errorf("")
}

func main() {
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
				Name:  "log-level",
				Usage: "Log level (trace, debug, info, warn, error, fatal)",
				Value: "",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output format (json, table)",
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
			&cli.BoolFlag{
				Name:  "version",
				Usage: "Prints the version of infracost and terraform",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "show-skipped",
				Usage: "Prints the list of free and unsupported resources",
				Value: false,
			},
		},
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			return customError(c, err.Error())
		},
		Action: func(c *cli.Context) error {
			config.Config.NoColor = c.Bool("no-color")
			color.NoColor = c.Bool("no-color")

			if c.Bool("version") {
				fmt.Println("Infracost", version.Version)
				v, err := terraform.TerraformVersion()
				fmt.Println(v)
				return err
			}

			err := config.Config.SetLogLevel(c.String("log-level"))
			if err != nil {
				return customError(c, err.Error())
			}

			if c.String("api-url") != "" {
				config.Config.ApiUrl = c.String("api-url")
			}

			provider := terraform.New()
			if err := provider.ProcessArgs(c); err != nil {
				return customError(c, err.Error())
			}

			s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))

			if !config.Config.IsLogging() {
				s.Suffix = " Calculating costs…"
				if !c.Bool("no-color") {
					_ = s.Color("fgHiGreen", "bold")
				}
				s.Start()
			} else {
				log.Info("Calculating costs…")
			}

			resources, err := provider.LoadResources()
			if err != nil {
				return errors.Wrap(err, "error loading resources")
			}

			if err := prices.PopulatePrices(resources); err != nil {
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

			s.Stop()

			if err != nil {
				return errors.Wrap(err, "output error")
			}

			fmt.Println(string(out))

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
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
