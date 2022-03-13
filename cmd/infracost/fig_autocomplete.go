package main

import (
	"github.com/spf13/cobra"
	genFigSpec "github.com/withfig/autocomplete-tools/packages/cobra"
)

func figAutocompleteCmd() *cobra.Command {
	opts := genFigSpec.Opts{
		Use: "fig-autocomplete",
	}
	// command hidden by default
	cmd := genFigSpec.NewCmdGenFigSpec(opts)
	cmd.Aliases = []string{}
	return cmd
}
