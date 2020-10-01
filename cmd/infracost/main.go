package main

import (
	"fmt"
	"os"

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
	err := config.Config.SetLogLevel(c.String("log-level"))
	if err != nil {
		return customError(c, err.Error(), true)
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
		Name:  "infracost",
		Usage: "Generate cost reports from Terraform plans",
		UsageText: `infracost [global options] command [command options] [arguments...]

Example:
	# Run infracost with a terraform directory
	infracost --tfdir /path/to/code

	# Run infracost with a JSON terraform plan file
	infracost --tfjson /path/to/plan/file

	# Run infracost with a terraform directory and a plan file in it
	infracost --tfdir /path/to/code --tfplan plan_file`,
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
