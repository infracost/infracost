package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/infracost/infracost/internal/providers/terraform"
	"github.com/infracost/infracost/internal/spin"
	"github.com/infracost/infracost/pkg/config"
	"github.com/infracost/infracost/pkg/version"
	log "github.com/sirupsen/logrus"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func usageError(c *cli.Context, msg string) {
	fmt.Fprintln(os.Stderr, color.HiRedString(msg)+"\n")
	c.App.Writer = os.Stderr
	cli.ShowAppHelpAndExit(c, 1)
}

var spinner *spin.Spinner

func handleGlobalFlags(c *cli.Context) error {
	config.Config.NoColor = c.Bool("no-color")
	color.NoColor = c.Bool("no-color")

	if c.IsSet("log-level") {
		err := config.Config.SetLogLevel(c.String("log-level"))
		if err != nil {
			usageError(c, err.Error())
		}
	}

	if c.String("pricing-api-endpoint") != "" {
		config.Config.PricingAPIEndpoint = c.String("pricing-api-endpoint")
	}

	return nil
}

func versionOutput(app *cli.App) string {
	s := fmt.Sprintf("Infracost %s", app.Version)
	v, err := terraform.TerraformVersion()
	if err != nil {
		log.Warnf("error determining Terraform version")
	} else {
		s += fmt.Sprintf("\n%s", v)
	}
	return s
}

func checkApiKey() bool {
	infracostApiKey := config.Config.ApiKey
	if config.Config.PricingAPIEndpoint == config.Config.DefaultPricingAPIEndpoint && infracostApiKey == "" {
		red := color.New(color.FgHiRed)
		bold := color.New(color.Bold, color.FgHiWhite)
		fmt.Fprintln(os.Stderr, red.Sprint("No INFRACOST_API_KEY environment variable is set."))
		fmt.Fprintln(os.Stderr, red.Sprintf("We run a free hosted API for cloud prices, to get an API key run"), bold.Sprint("`infracost register`\n"))
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
		fmt.Println(versionOutput(c.App))
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
			usageError(c, err.Error())
			return nil
		},
		Before:   handleGlobalFlags,
		Commands: []*cli.Command{registerCmd()},
		Action:   defaultCmd.Action,
	}

	defer func() {
		if err := recover(); err != nil {
			red := color.New(color.FgHiRed)
			bold := color.New(color.Bold, color.FgHiWhite)
			fmt.Fprintln(os.Stderr, red.Sprint("An unexpected error occurred\n"))
			fmt.Fprintf(os.Stderr, "%s\n%s\n", err, string(debug.Stack()))
			fmt.Fprintf(os.Stderr, "Environment:\n%s\n", versionOutput(app))
			fmt.Fprintln(os.Stderr, red.Sprint("\nPlease copy the above output and create a new issue at"), bold.Sprint("https://github.com/infracost/infracost/issues/new"))
		}
	}()
	if err := app.Run(os.Args); err != nil {
		if spinner != nil {
			spinner.Fail()
		}
		if err.Error() != "" {
			fmt.Fprintln(os.Stderr, color.HiRedString(err.Error()))
		}
		os.Exit(1)
	}
}
