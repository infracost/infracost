package ui

import (
	"fmt"
	"io"
	"os"

	"github.com/infracost/infracost/internal/version"
	"github.com/spf13/cobra"
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

func PrintUsageErrorAndExit(cmd *cobra.Command, msg string) {
	cmd.SetOut(os.Stderr)
	_ = cmd.Help()
	fmt.Fprintln(os.Stderr, "")
	PrintError(os.Stderr, msg)
	os.Exit(1)
}

func PrintUnexpectedError(err interface{}, stack string) {
	msg := fmt.Sprintf("\n%s %s\n\n%s\n%s\nEnvironment:\n%s\n\n%s %s\n",
		ErrorString("Error:"),
		"An unexpected error occurred",
		err,
		stack,
		fmt.Sprintf("Infracost %s", version.Version),
		"Please copy the above output and create a new issue at",
		LinkString("https://github.com/infracost/infracost/issues/new"),
	)

	fmt.Fprint(os.Stderr, msg)
}
