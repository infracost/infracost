package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// enum implements the flag.Value interface to provide a custom flag that validates
// the user inputs against a set of Allowed strings. This Allowed list is also
// used to generate cli autocompletion.
type enum struct {
	Allowed []string
	Value   string
}

func (e *enum) Type() string   { return "string" }
func (e *enum) String() string { return e.Value }

// Set changes the Value of enum to p. If p is not in the Allowed slice
// This function will error, which will cause a top level flag.Parse error.
// This halts Infracost execution early and results in an CLI output similar to the below:
//
// invalid argument "ds" for "--format" flag: valid arguments are "json", "table", "html".
func (e *enum) Set(p string) error {
	if !contains(e.Allowed, p) {
		b := strings.Builder{}
		for _, s := range e.Allowed {
			b.WriteString(fmt.Sprintf("%q, ", s))
		}

		return fmt.Errorf("valid arguments are %s", strings.TrimRight(b.String(), ", "))
	}

	e.Value = p

	return nil
}

func (e *enum) completion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return e.Allowed, cobra.ShellCompDirectiveDefault
}

func newEnumFlag(cmd *cobra.Command, name string, value string, usage string, allowed []string) {
	e := enum{
		Allowed: allowed,
		Value:   value,
	}

	cmd.Flags().Var(&e, name, fmt.Sprintf("%s: %s", usage, strings.Join(e.Allowed, ", ")))
	_ = cmd.RegisterFlagCompletionFunc(name, e.completion)
}
