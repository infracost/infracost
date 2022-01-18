package ui

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/version"
)

func PrintSuccess(w io.Writer, msg string) {
	fmt.Fprintf(w, "%s %s\n", SuccessString("Success:"), msg)
}

func PrintSuccessf(w io.Writer, msg string, a ...interface{}) {
	PrintSuccess(w, fmt.Sprintf(msg, a...))
}

func PrintError(w io.Writer, msg string) {
	fmt.Fprintf(w, "%s %s\n", ErrorString("Error:"), msg)
}

func PrintErrorf(w io.Writer, msg string, a ...interface{}) {
	PrintError(w, fmt.Sprintf(msg, a...))
}

func PrintWarning(w io.Writer, msg string) {
	fmt.Fprintf(w, "%s %s\n", WarningString("Warning:"), msg)
}

func PrintWarningf(w io.Writer, msg string, a ...interface{}) {
	PrintWarning(w, fmt.Sprintf(msg, a...))
}

func PrintUsage(cmd *cobra.Command) {
	cmd.SetOut(cmd.ErrOrStderr())
	_ = cmd.Help()
	cmd.Println("")
}

var (
	githubIssuesLink = LinkString("https://github.com/infracost/infracost/issues/new")

	unexpectedErrorMsg = "An unexpected error occurred. We have been notified of this issue and will look into is ASAP. If you would like to follow-up, you can create a GitHub issue at"
	stackErrorMsg      = "We have been notified of this issue and will look into is ASAP. Please copy the above output and create a new issue at:"
)

// PrintUnexpectedError prints a friendly user message to stdError.
// This should be used only in case of fatal/unexpected errors.
func PrintUnexpectedError(out io.Writer) {
	msg := fmt.Sprintf("\nEnvironment:\n%s\n\n%s %s\n",
		fmt.Sprintf("Infracost %s", version.Version),
		unexpectedErrorMsg,
		githubIssuesLink,
	)

	fmt.Fprint(out, msg)
}

// PrintUnexpectedErrorStack prints a full stack trace of a fatal error.
//
// In most cases PrintUnexpectedError function is suitable.
// This function should be called if the log level is debug and users wish to get additional info.
func PrintUnexpectedErrorStack(out io.Writer, err interface{}, stack string) {
	msg := fmt.Sprintf("\n%s %s\n\n%s\n%s\nEnvironment:\n%s\n\n%s %s\n",
		ErrorString("Error:"),
		"An unexpected error occurred",
		err,
		stack,
		fmt.Sprintf("Infracost %s", version.Version),
		stackErrorMsg,
		githubIssuesLink,
	)

	fmt.Fprint(out, msg)
}
