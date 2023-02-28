package main

import (
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/config/template"
	"github.com/infracost/infracost/internal/ui"
)

type generateConfigCommand struct {
	projectPath  string
	templatePath string
	outFile      string
}

func newGenerateConfigCommand() *cobra.Command {
	var gen generateConfigCommand

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Generate Infracost config file from a template file",
		Long:  "Generate Infracost config file from a template file",
		Example: `
      infracost generate config --project-path . --template-path infracost.yml.tmpl
      `,
		ValidArgs: []string{"--", "-"},
		RunE:      gen.run,
	}

	cmd.Flags().StringVar(&gen.projectPath, "project-path", ".", "Path to the Terraform project you want to run the template file on")
	cmd.Flags().StringVar(&gen.templatePath, "template-path", "", "Path to the Infracost template file which will generate the yml output")
	cmd.Flags().StringVar(&gen.outFile, "out-file", "", "Save output to a file")

	return cmd
}

func (g *generateConfigCommand) run(cmd *cobra.Command, args []string) error {
	if g.templatePath == "" {
		ui.PrintErrorf(cmd.ErrOrStderr(), "Please provide a path to an Infracost config template file.\n")
		ui.PrintUsage(cmd)
		return nil
	}

	if _, err := os.Stat(g.templatePath); err != nil {
		ui.PrintErrorf(cmd.ErrOrStderr(), "Provided template file %q does not exist\n", g.templatePath)
		ui.PrintUsage(cmd)
		return nil
	}

	projectPath, _ := os.Getwd()
	if g.projectPath != "." && g.projectPath != "" {
		projectPath = g.projectPath
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

	parser := template.NewParser(projectPath)
	err := parser.Compile(g.templatePath, out)
	if err != nil {
		ui.PrintErrorf(cmd.ErrOrStderr(), "Could not compile template file %q error: %s", g.templatePath, err)
	}

	return nil
}

func newGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate configuration to help run Infracost",
		Long:  "Generate configuration to help run Infracost",
		Example: ` Generate Infracost config file from a template file:

      infracost generate config --project-path . --template-path infracost.yml.tmpl
      `,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newGenerateConfigCommand())

	return cmd
}
