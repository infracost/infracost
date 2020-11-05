package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/providers/terraform"
	"github.com/infracost/infracost/internal/spin"
	"github.com/infracost/infracost/internal/update"
	"github.com/infracost/infracost/internal/version"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

var spinner *spin.Spinner

func usageError(c *cli.Context, msg string) {
	fmt.Fprintln(os.Stderr, color.HiRedString(msg)+"\n")
	c.App.Writer = os.Stderr
	cli.ShowAppHelpAndExit(c, 1)
}

func handleGlobalFlags(c *cli.Context) error {
	if c.IsSet("no-color") {
		config.Config.NoColor = c.Bool("no-color")
	}
	color.NoColor = config.Config.NoColor

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

func startUpdateCheck(c chan *update.Info) {
	go func() {
		updateInfo, err := update.CheckForUpdate()
		if err != nil {
			log.Debugf("error checking for update: %v", err)
		}
		c <- updateInfo
		close(c)
	}()
}

func versionOutput(app *cli.App) string {
	s := fmt.Sprintf("Infracost %s", app.Version)
	v, err := terraform.Version()

	if err != nil {
		log.Warnf("error determining Terraform version")
	} else {
		s += fmt.Sprintf("\n%s", v)
	}

	return s
}

func checkAPIKey() error {
	infracostAPIKey := config.Config.APIKey
	if config.Config.PricingAPIEndpoint == config.Config.DefaultPricingAPIEndpoint && infracostAPIKey == "" {
		red := color.New(color.FgHiRed)
		bold := color.New(color.Bold, color.FgHiWhite)

		return errors.New(fmt.Sprintf("%s\n%s %s",
			red.Sprint("No INFRACOST_API_KEY environment variable is set."),
			red.Sprintf("We run a free hosted API for cloud prices, to get an API key run"),
			bold.Sprint("`infracost register`"),
		))
	}

	return nil
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

	updateMessageChan := make(chan *update.Info)
	startUpdateCheck(updateMessageChan)

	app := &cli.App{
		Name:  "infracost",
		Usage: "Generate cost reports from Terraform plans",
		UsageText: `infracost [global options] command [command options] [arguments...]

Example:
	# Run infracost with a Terraform directory and var file
	infracost --tfdir /path/to/code --tfflags "-var-file=myvars.tfvars"

	# Run infracost against a Terraform state
	infracost --tfdir /path/to/code --use-tfstate

	# Run infracost with a JSON Terraform plan file
	infracost --tfjson /path/to/plan.json`,
		EnableBashCompletion: true,
		Version:              version.Version,
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:  "log-level",
				Usage: "Log level (trace, debug, info, warn, error, fatal)",
			},
			&cli.BoolFlag{
				Name:  "no-color",
				Usage: "Turn off colored output",
			},
			&cli.StringFlag{
				Name:  "pricing-api-endpoint",
				Usage: "Specify an alternate price list API URL",
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

	var appErr error

	defer func() {
		if appErr != nil {
			if spinner != nil {
				spinner.Fail()
			}

			if appErr.Error() != "" {
				fmt.Fprintf(os.Stderr, "%s\n", color.HiRedString(appErr.Error()))
			}
		}

		unexpectedErr := recover()
		if unexpectedErr != nil {
			if spinner != nil {
				spinner.Fail()
			}

			red := color.New(color.FgHiRed)
			bold := color.New(color.Bold, color.FgHiWhite)

			msg := fmt.Sprintf("\n%s\n%s\n%s\nEnvironment:\n%s\n\n%s %s\n",
				red.Sprint("An unexpected error occurred"),
				unexpectedErr,
				string(debug.Stack()),
				versionOutput(app),
				red.Sprint("Please copy the above output and create a new issue at"),
				bold.Sprint("https://github.com/infracost/infracost/issues/new"),
			)
			fmt.Fprint(os.Stderr, msg)
		}

		updateInfo := <-updateMessageChan
		if updateInfo != nil {
			msg := fmt.Sprintf("\n%s %s → %s\n%s\n",
				color.YellowString("A new version of Infracost is available:"),
				color.CyanString(version.Version),
				color.CyanString(updateInfo.LatestVersion),
				indent(color.YellowString(updateInfo.Cmd), "  "),
			)
			fmt.Fprint(os.Stderr, msg)
		}

		if appErr != nil || unexpectedErr != nil {
			os.Exit(1)
		}
	}()

	appErr = app.Run(os.Args)
}

func indent(s, indent string) string {
	lines := make([]string, 0)
	for _, j := range strings.Split(s, "\n") {
		lines = append(lines, indent+j)
	}

	return strings.Join(lines, "\n")
}
