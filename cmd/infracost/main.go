package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/ui"
	"github.com/infracost/infracost/internal/update"
	"github.com/infracost/infracost/internal/version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/fatih/color"
)

var spinner *ui.Spinner

func main() {
	var appErr error
	updateMessageChan := make(chan *update.Info)

	ctx, err := config.NewRunContextFromEnv(context.Background())
	if err != nil {
		if err.Error() != "" {
			ui.PrintError(os.Stderr, err.Error())
		}
		os.Exit(1)
	}

	defer func() {
		if appErr != nil {
			handleCLIError(ctx, appErr)
		}

		unexpectedErr := recover()
		if unexpectedErr != nil {
			handleUnexpectedErr(ctx, unexpectedErr)
		}

		handleUpdateMessage(updateMessageChan)

		if appErr != nil || unexpectedErr != nil {
			os.Exit(1)
		}
	}()

	startUpdateCheck(ctx, updateMessageChan)

	rootCmd := NewRootCommand(ctx)
	appErr = rootCmd.Execute()
}

func NewRootCommand(ctx *config.RunContext) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "infracost",
		Version: version.Version,
		Short:   "Cloud cost estimates for Terraform",
		Long: fmt.Sprintf(`Infracost - cloud cost estimates for Terraform

%s
  https://infracost.io/docs`, ui.BoldString("DOCS")),
		Example: `  Generate a cost diff from Terraform directory with any required Terraform flags:

      infracost diff --path /path/to/code --terraform-plan-flags "-var-file=my.tfvars"
	
  Generate a full cost breakdown from Terraform directory with any required Terraform flags:

      infracost breakdown --path /path/to/code --terraform-plan-flags "-var-file=my.tfvars"`,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			ctx.SetContextValue("command", cmd.Name())

			return loadGlobalFlags(ctx, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show the help
			return cmd.Help()
		},
	}

	rootCmd.PersistentFlags().Bool("no-color", false, "Turn off colored output")
	rootCmd.PersistentFlags().String("log-level", "", "Log level (trace, debug, info, warn, error, fatal)")

	rootCmd.AddCommand(registerCmd(ctx))
	rootCmd.AddCommand(configureCmd(ctx))
	rootCmd.AddCommand(diffCmd(ctx))
	rootCmd.AddCommand(breakdownCmd(ctx))
	rootCmd.AddCommand(outputCmd(ctx))
	rootCmd.AddCommand(completionCmd())

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

	return rootCmd
}

func startUpdateCheck(ctx *config.RunContext, c chan *update.Info) {
	go func() {
		updateInfo, err := update.CheckForUpdate(ctx)
		if err != nil {
			log.Debugf("error checking for update: %v", err)
		}
		c <- updateInfo
		close(c)
	}()
}

func checkAPIKey(apiKey string, apiEndpoint string, defaultEndpoint string) error {
	if apiEndpoint == defaultEndpoint && apiKey == "" {
		return fmt.Errorf(
			"No INFRACOST_API_KEY environment variable is set.\nWe run a free Cloud Pricing API, to get an API key run %s",
			ui.PrimaryString("infracost register"),
		)
	}

	return nil
}

func handleCLIError(ctx *config.RunContext, cliErr error) {
	if spinner != nil {
		spinner.Fail()
		fmt.Fprintln(os.Stderr, "")
	}

	if cliErr.Error() != "" {
		ui.PrintError(os.Stderr, cliErr.Error())
	}

	err := apiclient.ReportCLIError(ctx, cliErr)
	if err != nil {
		log.Warnf("Error reporting CLI error: %s", err)
	}
}

func handleUnexpectedErr(ctx *config.RunContext, unexpectedErr interface{}) {
	if spinner != nil {
		spinner.Fail()
		fmt.Fprintln(os.Stderr, "")
	}

	stack := string(debug.Stack())

	ui.PrintUnexpectedError(unexpectedErr, stack)

	err := apiclient.ReportCLIError(ctx, fmt.Errorf("%s\n%s", unexpectedErr, stack))
	if err != nil {
		log.Warnf("Error reporting unexpected error: %s", err)
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
	if cmd.Flags().Changed("no-color") {
		ctx.Config.NoColor, _ = cmd.Flags().GetBool("no-color")
	}
	color.NoColor = ctx.Config.NoColor

	if cmd.Flags().Changed("log-level") {
		ctx.Config.LogLevel, _ = cmd.Flags().GetString("log-level")
		err := ctx.Config.ConfigureLogger()
		if err != nil {
			return err
		}
	}

	ctx.SetContextValue("isDefaultPricingAPIEndpoint", ctx.Config.PricingAPIEndpoint == ctx.Config.DefaultPricingAPIEndpoint)

	flagNames := make([]string, 0)

	cmd.Flags().Visit(func(f *pflag.Flag) {
		flagNames = append(flagNames, f.Name)
	})

	ctx.SetContextValue("flags", flagNames)

	return nil
}
