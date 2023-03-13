package main

import (
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/config/template"
	"github.com/infracost/infracost/internal/ui"
)

type generateConfigCommand struct {
	repoPath     string
	templatePath string
	template     string
	outFile      string
}

func newGenerateConfigCommand() *cobra.Command {
	var gen generateConfigCommand

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Generate Infracost config file from a template file",
		Long:  "Generate Infracost config file from a template file",
		Example: `
      infracost generate config --repo-path . --template-path infracost.yml.tmpl
      `,
		ValidArgs: []string{"--", "-"},
		RunE:      gen.run,
	}

	cmd.Flags().StringVar(&gen.repoPath, "repo-path", ".", "Path to the Terraform repo or directory you want to run the template file on")
	cmd.Flags().StringVar(&gen.template, "template", "", "Infracost template string that will generate the config-file yaml output")
	cmd.Flags().StringVar(&gen.templatePath, "template-path", "", "Path to the Infracost template file that will generate the config-file yaml output")
	cmd.Flags().StringVar(&gen.outFile, "out-file", "", "Save output to a file")

	return cmd
}

func (g *generateConfigCommand) run(cmd *cobra.Command, args []string) error {
	if g.templatePath == "" && g.template == "" {
		ui.PrintErrorf(cmd.ErrOrStderr(), "Please provide an Infracost config template.\n")
		ui.PrintUsage(cmd)
		return nil
	}

	if _, err := os.Stat(g.templatePath); g.templatePath != "" && err != nil {
		ui.PrintErrorf(cmd.ErrOrStderr(), "Provided template file %q does not exist\n", g.templatePath)
		ui.PrintUsage(cmd)
		return nil
	}

	repoPath, _ := os.Getwd()
	if g.repoPath != "." && g.repoPath != "" {
		repoPath = g.repoPath
	}

	var out io.Writer = cmd.OutOrStderr()
	if g.outFile != "" {
		var err error
		out, err = os.Create(g.outFile)
		if err != nil {
			ui.PrintErrorf(cmd.ErrOrStderr(), "Could not create out file %s %s", g.outFile, err)
			return nil
		}
	}

	parser := template.NewParser(repoPath)
	if g.template != "" {
		err := parser.Compile(g.template, out)
		if err != nil {
			ui.PrintErrorf(cmd.ErrOrStderr(), "Could not compile template error: %s", err)
		}

		return nil
	}

	err := parser.CompileFromFile(g.templatePath, out)
	if err != nil {
		ui.PrintErrorf(cmd.ErrOrStderr(), "Could not compile template error: %s", err)
	}
	return nil
}

func newGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate configuration to help run Infracost",
		Long:  "Generate configuration to help run Infracost",
		Example: ` Generate Infracost config file from a template file:

      infracost generate config --repo-path . --template-path infracost.yml.tmpl
      `,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newGenerateConfigCommand())

	return cmd
}
