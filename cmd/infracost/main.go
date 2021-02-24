package main

import (
	"fmt"
	"os"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/events"
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

func usageWarning(msg string) {
	fmt.Fprintln(os.Stderr, color.YellowString(msg)+"\n")
}

func main() {
	var appErr error
	updateMessageChan := make(chan *update.Info)

	cfg := config.DefaultConfig()
	appErr = cfg.LoadFromEnv()

	var app *cli.App

	defer func() {
		if appErr != nil {
			handleAppErr(cfg, appErr)
		}

		unexpectedErr := recover()
		if unexpectedErr != nil {
			handleUnexpectedErr(cfg, app, unexpectedErr)
		}

		handleUpdateMessage(updateMessageChan)

		if appErr != nil || unexpectedErr != nil {
			os.Exit(1)
		}
	}()

	cli.VersionFlag = &cli.BoolFlag{
		Name:  "version",
		Usage: "Prints the version of infracost and terraform",
	}

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println(versionOutput(c.App))
	}

	startUpdateCheck(cfg, updateMessageChan)

	defaultCmd := breakdownCmd(cfg)

	app = &cli.App{
		Name:  "infracost",
		Usage: "Generate cost estimates from Terraform",
		UsageText: `infracost [global options] command [command options] [arguments...]

USAGE METHODS:
	# 1. Use terraform directory with any required terraform flags
	infracost --terraform-dir /path/to/code --terraform-plan-flags "-var-file=myvars.tfvars"

	# 2. Use terraform state file
	infracost --terraform-dir /path/to/code --terraform-use-state

	# 3. Use terraform plan JSON
	terraform plan -out plan.save .
	terraform show -json plan.save > plan.json
	infracost --terraform-json-file /path/to/plan.json

	# 4. Use terraform plan file, relative to terraform-dir
	terraform plan -out plan.save .
	infracost --terraform-dir /path/to/code --terraform-plan-file plan.save

DOCS: https://infracost.io/docs`,
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
				Usage: "Specify an alternate Cloud Pricing API URL",
			},
		}, defaultCmd.Flags...),
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			usageError(c, err.Error())
			return nil
		},
		Before: func(c *cli.Context) error {
			return loadGlobalFlags(cfg, c)
		},
		Commands: []*cli.Command{registerCmd(cfg), diffCmd(cfg), breakdownCmd(cfg), reportCmd(cfg)},
		Action:   defaultCmd.Action,
	}

	appErr = app.Run(os.Args)
}

func startUpdateCheck(cfg *config.Config, c chan *update.Info) {
	go func() {
		updateInfo, err := update.CheckForUpdate(cfg)
		if err != nil {
			log.Debugf("error checking for update: %v", err)
		}
		c <- updateInfo
		close(c)
	}()
}

func versionOutput(app *cli.App) string {
	return fmt.Sprintf("Infracost %s", app.Version)
}

func checkAPIKey(apiKey string, apiEndpoint string, defaultEndpoint string) error {
	if apiEndpoint == defaultEndpoint && apiKey == "" {
		red := color.New(color.FgHiRed)
		bold := color.New(color.Bold)

		return errors.New(fmt.Sprintf("%s\n%s %s",
			red.Sprint("No INFRACOST_API_KEY environment variable is set."),
			red.Sprintf("We run a free Cloud Pricing API, to get an API key run"),
			bold.Sprint("`infracost register`"),
		))
	}

	return nil
}

func handleAppErr(cfg *config.Config, err error) {
	if spinner != nil {
		spinner.Fail()
	}

	if err.Error() != "" {
		fmt.Fprintf(os.Stderr, "%s\n", color.HiRedString(err.Error()))
	}

	msg := stripColor(err.Error())
	var eventsError *events.Error
	if errors.As(err, &eventsError) {
		msg = stripColor(eventsError.Label)
	}
	events.SendReport(cfg, "error", msg)
}

func handleUnexpectedErr(cfg *config.Config, app *cli.App, unexpectedErr interface{}) {
	if spinner != nil {
		spinner.Fail()
	}

	red := color.New(color.FgHiRed)
	bold := color.New(color.Bold)
	stack := string(debug.Stack())

	v := ""
	if app != nil {
		v = versionOutput(app)
	}

	msg := fmt.Sprintf("\n%s\n%s\n%s\nEnvironment:\n%s\n\n%s %s\n",
		red.Sprint("An unexpected error occurred"),
		unexpectedErr,
		stack,
		v,
		red.Sprint("Please copy the above output and create a new issue at"),
		bold.Sprint("https://github.com/infracost/infracost/issues/new"),
	)
	fmt.Fprint(os.Stderr, msg)

	events.SendReport(cfg, "error", fmt.Sprintf("%s\n%s", unexpectedErr, stack))
}

func handleUpdateMessage(updateMessageChan chan *update.Info) {
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
}

func loadGlobalFlags(cfg *config.Config, c *cli.Context) error {
	if c.IsSet("no-color") {
		cfg.NoColor = c.Bool("no-color")
	}
	color.NoColor = cfg.NoColor

	if c.IsSet("log-level") {
		cfg.LogLevel = c.String("log-level")
		err := cfg.ConfigureLogger()
		if err != nil {
			return err
		}
	}

	if c.IsSet("pricing-api-endpoint") {
		cfg.PricingAPIEndpoint = c.String("pricing-api-endpoint")
	}

	cfg.Environment.IsDefaultPricingAPIEndpoint = cfg.PricingAPIEndpoint == cfg.DefaultPricingAPIEndpoint
	cfg.Environment.Flags = c.FlagNames()

	return nil
}

func indent(s, indent string) string {
	lines := make([]string, 0)
	for _, j := range strings.Split(s, "\n") {
		lines = append(lines, indent+j)
	}

	return strings.Join(lines, "\n")
}

func stripColor(str string) string {
	ansi := "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	re := regexp.MustCompile(ansi)
	return re.ReplaceAllString(str, "")
}
