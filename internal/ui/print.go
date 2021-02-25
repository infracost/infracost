package ui

import (
	"fmt"
	"os"

	"github.com/infracost/infracost/internal/version"
	"github.com/spf13/cobra"
)

func PrintSuccess(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", SuccessString("Success:"), msg)
}

func PrintSuccessf(msg string, a ...interface{}) {
	PrintSuccess(fmt.Sprintf(msg, a...))
}

func PrintError(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", ErrorString("Error:"), msg)
}

func PrintErrorf(msg string, a ...interface{}) {
	PrintError(fmt.Sprintf(msg, a...))
}

func PrintWarning(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", WarningString("Warning:"), msg)
}

func PrintWarningf(msg string, a ...interface{}) {
	PrintWarning(fmt.Sprintf(msg, a...))
}

func PrintUsageError(cmd *cobra.Command, msg string) {
	fmt.Fprintln(os.Stderr, ErrorString(msg)+"\n")
	cmd.SetOut(os.Stderr)
	_ = cmd.Help()
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
