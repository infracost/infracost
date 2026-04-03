package ui

import (
	"fmt"
	"io"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/version"
)

// WriteWarningFunc defines an interface that writes the provided msg as a warning
// to an underlying writer.
type WriteWarningFunc func(msg string)

func PrintError(w io.Writer, msg string) {
	fmt.Fprintf(w, "%s %s\n", ErrorString("Error:"), msg)
}

func PrintErrorf(w io.Writer, msg string, a ...any) {
	PrintError(w, fmt.Sprintf(msg, a...))
}

func PrintUsage(cmd *cobra.Command) {
	cmd.SetOut(cmd.ErrOrStderr())
	_ = cmd.Help()
	cmd.Println("")
}

var (
	githubIssuesLink = LinkString("https://github.com/infracost/infracost/issues/new")

	stackErrorMsg = "An unexpected error occurred. We've been notified of it and will investigate it soon. If you would like to follow-up, please copy the above output and create an issue at:"
)

// PrintUnexpectedErrorStack prints a full stack trace of a fatal error.
func PrintUnexpectedErrorStack(err error) {
	logging.Logger.Error().Msgf("%s\n\n%s\nEnvironment:\n%s\n\n%s %s\n",
		"An unexpected error occurred",
		err,
		fmt.Sprintf("Infracost %s", version.Version),
		stackErrorMsg,
		githubIssuesLink,
	)
}

func ProjectDisplayName(ctx *config.RunContext, name string) string {
	return FormatIfNotCI(ctx, func(s string) string {
		return color.BlueString(BoldString(s))
	}, name)
}

func DirectoryDisplayName(ctx *config.RunContext, name string) string {
	return FormatIfNotCI(ctx, UnderlineString, name)
}
