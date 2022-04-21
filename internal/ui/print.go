package ui

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/version"
)

// WriteWarningFunc defines an interface that writes the provided msg as a warning
// to an underlying writer.
type WriteWarningFunc func(msg string)

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

	stackErrorMsg = "An unexpected error occurred. We've been notified of it and will investigate it soon. If you would like to follow-up, please copy the above output and create an issue at:"
)

// PrintUnexpectedErrorStack prints a full stack trace of a fatal error.
func PrintUnexpectedErrorStack(out io.Writer, err error) {
	msg := fmt.Sprintf("\n%s %s\n\n%s\nEnvironment:\n%s\n\n%s %s\n",
		ErrorString("Error:"),
		"An unexpected error occurred",
		err,
		fmt.Sprintf("Infracost %s", version.Version),
		stackErrorMsg,
		githubIssuesLink,
	)

	fmt.Fprint(out, msg)
}
