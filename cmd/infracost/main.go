package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	stdLog "log"
	"os"
	"runtime/debug"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/ui"
	"github.com/infracost/infracost/internal/update"
	"github.com/infracost/infracost/internal/version"
)

func init() {
	// set the stdlib default logger to flush to discard, this is done as a number of
	// Terraform libs use the std logger directly, which impacts Infracost output.
	stdLog.SetOutput(ioutil.Discard)
}

func main() {
	Run(nil, nil)
}

// Run starts the Infracost application with the configured cobra cmds.
// Cmd args and flags are parsed from the cli, but can also be directly injected
// using the modifyCtx and args parameters.
func Run(modifyCtx func(*config.RunContext), args *[]string) {
	ctx, err := config.NewRunContextFromEnv(context.Background())
	if err != nil {
		if err.Error() != "" {
			ui.PrintError(ctx.ErrWriter, err.Error())
		}

		ctx.Exit(1)
	}

	if modifyCtx != nil {
		modifyCtx(ctx)
	}

	var appErr error
	updateMessageChan := make(chan *update.Info)

	defer func() {
		if appErr != nil {
			if v, ok := appErr.(*clierror.PanicError); ok {
				handleUnexpectedErr(ctx, v)
			} else {
				handleCLIError(ctx, appErr)
			}
		}

		unexpectedErr := recover()
		if unexpectedErr != nil {
			panicErr := clierror.NewPanicError(fmt.Errorf("%s", unexpectedErr), debug.Stack())
			handleUnexpectedErr(ctx, panicErr)
		}

		handleUpdateMessage(updateMessageChan)

		if appErr != nil || unexpectedErr != nil {
			ctx.Exit(1)
		}
	}()

	startUpdateCheck(ctx, updateMessageChan)

	rootCmd := newRootCmd(ctx)
	if args != nil {
		rootCmd.SetArgs(*args)
	}

	appErr = rootCmd.Execute()
}

type debugWriter struct {
	f *os.File
}

func (d debugWriter) Write(p []byte) (n int, err error) {
	p = bytes.Trim(p, " \n\t")
	return d.f.Write(append(p, []byte(",\n")...))
}

func newRootCmd(ctx *config.RunContext) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "infracost",
		Version: version.Version,
		Short:   "Cloud cost estimates for Terraform",
		Long: fmt.Sprintf(`Infracost - cloud cost estimates for Terraform

%s
  Quick start: https://infracost.io/docs
  Add cost estimates to your pull requests: https://infracost.io/cicd`, ui.BoldString("DOCS")),
		Example: `  Show cost diff from Terraform directory:

      infracost breakdown --path /code --format json --out-file infracost-base.json
      # Make Terraform code changes
      infracost diff --path /code --compare-to infracost-base.json

  Show cost breakdown from Terraform directory:

      infracost breakdown --path /code --terraform-var-file my.tfvars`,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			ctx.SetContextValue("command", cmd.Name())
			if cmd.Name() == "comment" || (cmd.Parent() != nil && cmd.Parent().Name() == "comment") {
				ctx.SetIsInfracostComment()
			}
			out, _ := cmd.Flags().GetBool("debug-report")
			if out {
				debugFile := "infracost-debug-report.json"
				var f *os.File
				var err error

				if _, serr := os.Stat(debugFile); serr != nil {
					f, err = os.OpenFile(debugFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
				} else {
					f, err = os.Create(debugFile)
				}

				if err != nil {
					return fmt.Errorf("could not generate debug report file %w", err)
				}
				_, _ = f.WriteString("[\n")

				writer := debugWriter{f: f}
				ctx.ErrWriter = writer
				ctx.Config.SetLogWriter(writer)
			}
			err := loadGlobalFlags(ctx, cmd)
			if err != nil {
				return err
			}

			loadCloudSettings(ctx)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show the help
			return cmd.Help()
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			out, _ := cmd.Flags().GetBool("debug-report")
			if out {
				if f, ok := ctx.Config.LogWriter().(debugWriter); ok {
					_, _ = f.f.WriteString("{\"msg\":\"program finished\"}\n")

					_, _ = f.f.WriteString("]")
					_ = f.f.Close()
				}
			}

			return nil
		},
	}

	rootCmd.PersistentFlags().Bool("no-color", false, "Turn off colored output")
	rootCmd.PersistentFlags().String("log-level", "", "Log level (trace, debug, info, warn, error, fatal)")
	rootCmd.PersistentFlags().Bool("debug-report", false, "Generate a debug report file which can be sent to Infracost team")

	rootCmd.AddCommand(authCmd(ctx))
	rootCmd.AddCommand(registerCmd(ctx))
	rootCmd.AddCommand(configureCmd(ctx))
	rootCmd.AddCommand(diffCmd(ctx))
	rootCmd.AddCommand(breakdownCmd(ctx))
	rootCmd.AddCommand(outputCmd(ctx))
	rootCmd.AddCommand(commentCmd(ctx))
	rootCmd.AddCommand(completionCmd())
	rootCmd.AddCommand(figAutocompleteCmd())

	rootCmd.SetUsageTemplate(fmt.Sprintf(`%s{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

%s
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

%s
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

%s{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

%s
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

%s
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

%s{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`,
		ui.BoldString("USAGE"),
		ui.BoldString("ALIAS"),
		ui.BoldString("EXAMPLES"),
		ui.BoldString("AVAILABLE COMMANDS"),
		ui.BoldString("FLAGS"),
		ui.BoldString("GLOBAL FLAGS"),
		ui.BoldString("ADDITIONAL HELP TOPICS"),
	))

	rootCmd.SetVersionTemplate("Infracost {{.Version}}\n")
	rootCmd.SetOut(ctx.OutWriter)
	rootCmd.SetErr(ctx.ErrWriter)

	return rootCmd
}

func startUpdateCheck(ctx *config.RunContext, c chan *update.Info) {
	go func() {
		updateInfo, err := update.CheckForUpdate(ctx)
		if err != nil {
			logging.Logger.WithError(err).Debug("error checking for Infracost CLI update")
		}
		c <- updateInfo
		close(c)
	}()
}

func loadCloudSettings(ctx *config.RunContext) {
	if ctx.Config.IsSelfHosted() || (ctx.Config.EnableCloud != nil && !*ctx.Config.EnableCloud) {
		return
	}

	dashboardClient := apiclient.NewDashboardAPIClient(ctx)
	result, err := dashboardClient.QueryCLISettings()
	if err != nil {
		logging.Logger.WithError(err).Debug("Failed to load settings from Infracost Cloud ")
		// ignore the error so the command can continue without failing
		return
	}
	logging.Logger.WithFields(log.Fields{"result": fmt.Sprintf("%+v", result)}).Debug("Successfully loaded settings from Infracost Cloud")

	ctx.Config.EnableCloudForOrganization = result.CloudEnabled
}

func checkAPIKey(apiKey string, apiEndpoint string, defaultEndpoint string) error {
	if apiEndpoint == defaultEndpoint && apiKey == "" {
		return fmt.Errorf(
			"No INFRACOST_API_KEY environment variable is set.\nWe run a free Cloud Pricing API, to get an API key run %s",
			ui.PrimaryString("infracost auth login"),
		)
	}

	return nil
}

func handleCLIError(ctx *config.RunContext, cliErr error) {
	if cliErr.Error() != "" {
		ui.PrintError(ctx.ErrWriter, cliErr.Error())
	}

	err := apiclient.ReportCLIError(ctx, cliErr, true)
	if err != nil {
		logging.Logger.WithError(err).Warn("error reporting CLI error")
	}
}

func handleUnexpectedErr(ctx *config.RunContext, err error) {
	ui.PrintUnexpectedErrorStack(ctx.ErrWriter, err)

	err = apiclient.ReportCLIError(ctx, err, false)
	if err != nil {
		logging.Logger.WithError(err).Warn("error sending unexpected runtime error")
	}
}

func handleUpdateMessage(updateMessageChan chan *update.Info) {
	updateInfo := <-updateMessageChan
	if updateInfo != nil {
		msg := fmt.Sprintf("\n%s %s %s → %s\n%s\n",
			ui.WarningString("Update:"),
			"A new version of Infracost is available:",
			ui.PrimaryString(version.Version),
			ui.PrimaryString(updateInfo.LatestVersion),
			ui.Indent(updateInfo.Cmd, "  "),
		)
		fmt.Fprint(os.Stderr, msg)
	}
}

func loadGlobalFlags(ctx *config.RunContext, cmd *cobra.Command) error {
	if ctx.IsCIRun() {
		ctx.Config.NoColor = true
	}
	if cmd.Flags().Changed("no-color") {
		ctx.Config.NoColor, _ = cmd.Flags().GetBool("no-color")
	}
	color.NoColor = ctx.Config.NoColor

	if cmd.Flags().Changed("log-level") {
		ctx.Config.LogLevel, _ = cmd.Flags().GetString("log-level")
		err := logging.ConfigureBaseLogger(ctx.Config)
		if err != nil {
			return err
		}
	}

	if cmd.Flags().Changed("debug-report") {
		ctx.Config.DebugReport, _ = cmd.Flags().GetBool("debug-report")
		err := logging.ConfigureBaseLogger(ctx.Config)
		if err != nil {
			return err
		}
	}

	ctx.SetContextValue("dashboardEnabled", ctx.Config.EnableDashboard)
	if ctx.Config.EnableCloud != nil {
		ctx.SetContextValue("cloudEnabled", ctx.Config.EnableCloud)
	}
	ctx.SetContextValue("isDefaultPricingAPIEndpoint", ctx.Config.PricingAPIEndpoint == ctx.Config.DefaultPricingAPIEndpoint)

	flagNames := make([]string, 0)

	cmd.Flags().Visit(func(f *pflag.Flag) {
		flagNames = append(flagNames, f.Name)
	})

	ctx.SetContextValue("flags", flagNames)

	return nil
}

// saveOutFile saves the output of the command to the file path past in the `--out-file` flag
func saveOutFile(ctx *config.RunContext, cmd *cobra.Command, outFile string, b []byte) error {
	return saveOutFileWithMsg(ctx, cmd, outFile, fmt.Sprintf("Output saved to %s", outFile), b)
}

// saveOutFile saves the output of the command to the file path past in the `--out-file` flag
func saveOutFileWithMsg(ctx *config.RunContext, cmd *cobra.Command, outFile, successMsg string, b []byte) error {
	err := ioutil.WriteFile(outFile, b, 0644) // nolint:gosec
	if err != nil {
		return errors.Wrap(err, "Unable to save output")
	}

	if ctx.Config.IsLogging() {
		logging.Logger.Info(successMsg)
	} else {
		cmd.PrintErrf("%s\n", successMsg)
	}

	return nil
}
