package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/briandowns/spinner"
	"github.com/infracost/infracost/internal/providers/terraform"
	"github.com/infracost/infracost/pkg/config"
	"github.com/infracost/infracost/pkg/version"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func customError(c *cli.Context, msg string, showHelp bool) error {
	color.HiRed(fmt.Sprintf("%v\n", msg))
	if showHelp {
		_ = cli.ShowAppHelp(c)
	}

	return fmt.Errorf("")
}

var calcSpinner *spinner.Spinner

func handleGlobalFlags(c *cli.Context) error {
	config.Config.NoColor = c.Bool("no-color")
	color.NoColor = c.Bool("no-color")

	switch strings.ToUpper(c.String("log-level")) {
	case "TRACE":
		log.SetLevel(log.TraceLevel)
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	if c.String("pricing-api-endpoint") != "" {
		config.Config.PricingAPIEndpoint = c.String("pricing-api-endpoint")
	}

	return nil
}

func checkApiKey() bool {
	infracostApiKey := config.Config.ApiKey
	if config.Config.PricingAPIEndpoint == config.Config.DefaultPricingAPIEndpoint && infracostApiKey == "" {
		color.Yellow("No INFRACOST_API_KEY environment variable is set.")
		c := color.New(color.Bold, color.FgHiWhite)
		fmt.Printf("We run a free hosted API for cloud prices, to get an API key run `%s`\n", c.Sprint("infracost register"))
		return false
	}
	return true
}

func main() {
	defaultCmd := defaultCmd()

	cli.VersionFlag = &cli.BoolFlag{
		Name:  "version",
		Usage: "Prints the version of infracost and terraform",
	}

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println("Infracost", c.App.Version)
		v, err := terraform.TerraformVersion()
		if err != nil {
			log.Warnf("error determining Terraform version")
		} else {
			fmt.Println(v)
		}
	}

	app := &cli.App{
		Name:                 "infracost",
		Usage:                "Generate cost reports from Terraform plans",
		EnableBashCompletion: true,
		Version:              version.Version,
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:  "log-level",
				Usage: "Log level (trace, debug, info, warn, error, fatal)",
				Value: "",
			},
			&cli.BoolFlag{
				Name:  "no-color",
				Usage: "Turn off colored output",
				Value: false,
			},
			&cli.StringFlag{
				Name:  "pricing-api-endpoint",
				Usage: "Specify an alternate price list API URL",
				Value: config.Config.PricingAPIEndpoint,
			},
		}, defaultCmd.Flags...),
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			return customError(c, err.Error(), true)
		},
		Before:   handleGlobalFlags,
		Commands: []*cli.Command{registerCmd()},
		Action:   defaultCmd.Action,
	}

	if err := app.Run(os.Args); err != nil {
		if calcSpinner != nil {
			calcSpinner.Stop()
		}
		if err.Error() != "" {
			color.HiRed(fmt.Sprintf("%v\n", err.Error()))
		}
		os.Exit(1)
	}
}
