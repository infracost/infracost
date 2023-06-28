package main

import (
	"bytes"
	"fmt"
	"github.com/infracost/infracost/internal/config"
	"gopkg.in/yaml.v2"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/config/template"
	"github.com/infracost/infracost/internal/ui"
	"github.com/infracost/infracost/internal/vcs"
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
		Long: `Generate Infracost config file from a template file. See docs for template examples and syntax:

  https://www.infracost.io/docs/features/config_file/#template-syntax`,
		Example: `
      infracost generate config --repo-path . --template-path infracost.yml.tmpl --out-file infracost.yml
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

	var buf bytes.Buffer

	m, err := vcs.MetadataFetcher.Get(repoPath, nil)
	if err != nil {
		ui.PrintWarningf(cmd.ErrOrStderr(), "could not fetch git metadata err: %s, default template variables will be blank", err)
	}

	variables := template.Variables{
		Branch: m.Branch.Name,
	}
	if m.PullRequest != nil {
		variables.BaseBranch = m.PullRequest.BaseBranch
	}

	parser := template.NewParser(repoPath, variables)
	if g.template != "" {
		err := parser.Compile(g.template, &buf)
		if err != nil {
			return err
		}
	} else {
		err := parser.CompileFromFile(g.templatePath, &buf)
		if err != nil {
			return err
		}
	}

	// Write the generated YAML
	var out io.Writer = cmd.OutOrStderr()
	if g.outFile != "" {
		var err error
		out, err = os.Create(g.outFile)
		if err != nil {
			return fmt.Errorf("could not create out file %s: %s", g.outFile, err)
		}
	}

	// save the contents of the buffer for validation since WriteTo drains the buffer
	bufStr := buf.String()

	_, err = buf.WriteTo(out)
	if err != nil {
		return fmt.Errorf("could not write file %s: %s", g.outFile, err)
	}

	cmd.Printf("\n")

	// Validate the generated YAML
	content := []byte(os.ExpandEnv(bufStr))

	var cfgFile config.ConfigFileSpec

	err = yaml.Unmarshal(content, &cfgFile)
	if err != nil {
		// we have to make this custom error type checking here
		// as indentations cause the yaml.Unmarshal to panic
		// it catches the panic and returns an error but in order
		// not to stutter the errors we should check here for
		// our custom error type.
		if _, ok := err.(*config.YamlError); ok {
			return fmt.Errorf("invalid config file: %s", err)
		}

		// if we receive a caught panic error, wrap the message in something more user-friendly
		return fmt.Errorf("could not validate generated config file, check file syntax: %s", err)
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
