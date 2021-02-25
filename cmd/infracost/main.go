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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/fatih/color"
)

var spinner *spin.Spinner

func usageError(cmd *cobra.Command, msg string) {
	fmt.Fprintln(os.Stderr, color.HiRedString(msg)+"\n")
	cmd.SetOut(os.Stderr)
	_ = cmd.Help()
}

func usageWarning(msg string) {
	fmt.Fprintln(os.Stderr, color.YellowString(msg)+"\n")
}

func main() {
	var appErr error
	updateMessageChan := make(chan *update.Info)

	cfg := config.DefaultConfig()
	appErr = cfg.LoadFromEnv()

	defer func() {
		if appErr != nil {
			handleAppErr(cfg, appErr)
		}

		unexpectedErr := recover()
		if unexpectedErr != nil {
			handleUnexpectedErr(cfg, unexpectedErr)
		}

		handleUpdateMessage(updateMessageChan)

		if appErr != nil || unexpectedErr != nil {
			os.Exit(1)
		}
	}()

	startUpdateCheck(cfg, updateMessageChan)

	rootCmd := &cobra.Command{
		Use:     "infracost",
		Version: version.Version,
		Short:   "Cloud cost estimates for Terraform",
		Long: `Infracost - cloud cost estimates for Terraform

Generate a cost diff from terraform directory with any required terraform flags:

  infracost diff --terraform-dir /path/to/code --terraform-plan-flags "-var-file=myvars.tfvars"

Generate a full cost breakdown from terraform directory with any required terraform flags:

  infracost breakdown --terraform-dir /path/to/code --terraform-plan-flags "-var-file=myvars.tfvars"

Docs:
  https://infracost.io/docs`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return loadGlobalFlags(cfg, cmd)
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			deprecationMsg := "The root command is deprecated and will be removed in v0.8.0. Please use `infracost breakdown`."
			usageWarning(deprecationMsg)

			handleDeprecatedEnvVars(deprecatedEnvVarMapping)
			handleDeprecatedFlags(cmd, deprecatedFlagsMapping)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return breakdownCmd(cfg).RunE(cmd, args)
		},
	}

	// Add the run flags and hide them since the root command is deprected
	addDeprecatedBreakdownFlags(rootCmd)
	addRunInputFlags(rootCmd)
	addRunOutputFlags(rootCmd)
	rootCmd.Flags().VisitAll(func(f *pflag.Flag) {
		f.Hidden = true
	})

	rootCmd.PersistentFlags().Bool("no-color", false, "Turn off colored output")
	rootCmd.PersistentFlags().String("log-level", "", "Log level (trace, debug, info, warn, error, fatal)")

	rootCmd.AddCommand(registerCmd(cfg))
	rootCmd.AddCommand(diffCmd(cfg))
	rootCmd.AddCommand(breakdownCmd(cfg))
	rootCmd.AddCommand(outputCmd(cfg))
	rootCmd.AddCommand(reportCmd(cfg))

	rootCmd.SetVersionTemplate("Infracost {{.Version}}\n")

	appErr = rootCmd.Execute()
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

func versionOutput() string {
	return fmt.Sprintf("Infracost %s", version.Version)
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

func handleUnexpectedErr(cfg *config.Config, unexpectedErr interface{}) {
	if spinner != nil {
		spinner.Fail()
	}

	red := color.New(color.FgHiRed)
	bold := color.New(color.Bold)
	stack := string(debug.Stack())

	msg := fmt.Sprintf("\n%s\n%s\n%s\nEnvironment:\n%s\n\n%s %s\n",
		red.Sprint("An unexpected error occurred"),
		unexpectedErr,
		stack,
		versionOutput(),
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

func loadGlobalFlags(cfg *config.Config, cmd *cobra.Command) error {
	if cmd.Flags().Changed("no-color") {
		cfg.NoColor, _ = cmd.Flags().GetBool("no-color")
	}
	color.NoColor = cfg.NoColor

	if cmd.Flags().Changed("log-level") {
		cfg.LogLevel, _ = cmd.Flags().GetString("log-level")
		err := cfg.ConfigureLogger()
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("pricing-api-endpoint") {
		cfg.PricingAPIEndpoint, _ = cmd.Flags().GetString("pricing-api-endpoint")
	}

	cfg.Environment.IsDefaultPricingAPIEndpoint = cfg.PricingAPIEndpoint == cfg.DefaultPricingAPIEndpoint

	flagNames := make([]string, 0)

	cmd.Flags().Visit(func(f *pflag.Flag) {
		flagNames = append(flagNames, f.Name)
	})

	cfg.Environment.Flags = flagNames

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
