package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/config/template"
	"github.com/infracost/infracost/internal/hcl"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/providers"
	"github.com/infracost/infracost/internal/ui"
	"github.com/infracost/infracost/internal/vcs"
)

type generateConfigCommand struct {
	wd           string
	templatePath string
	template     string
	outFile      string
	treeFile     string
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

	cmd.Flags().StringVar(&gen.wd, "repo-path", ".", "Path to the Terraform repo or directory you want to run the template file on")
	cmd.Flags().StringVar(&gen.template, "template", "", "Infracost template string that will generate the config-file yaml output")
	cmd.Flags().StringVar(&gen.templatePath, "template-path", "", "Path to the Infracost template file that will generate the config-file yaml output")
	cmd.Flags().StringVar(&gen.outFile, "out-file", "", "Save output to a file")
	cmd.Flags().StringVar(&gen.treeFile, "tree-file", "", "Save a simplified tree of the detected projects and var files to a file")

	return cmd
}

func (g *generateConfigCommand) run(cmd *cobra.Command, args []string) error {
	if _, err := os.Stat(g.templatePath); g.templatePath != "" && err != nil {
		ui.PrintErrorf(cmd.ErrOrStderr(), "Provided template file %q does not exist\n", g.templatePath)
		ui.PrintUsage(cmd)
		return nil
	}

	wd, _ := os.Getwd()
	if g.wd != "." && g.wd != "" {
		wd = g.wd
	}

	var buf bytes.Buffer

	ctx := config.EmptyRunContext()
	err := ctx.Config.LoadFromEnv()
	if err != nil {
		logging.Logger.Warn().Msgf("failed to load config from env: %s", err)
	}

	hasTemplate := g.template != "" || g.templatePath != ""
	var definedProjects bool
	if hasTemplate {
		var reader io.ReadSeeker
		if g.templatePath != "" {
			file, err := os.Open(g.templatePath)
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer file.Close()
			reader = file
		} else {
			reader = strings.NewReader(g.template)
		}

		partialConfig, err := unmarshalAutoDetectSection(reader)
		if err != nil {
			return fmt.Errorf("failed to unmarshal autodetect section: %w", err)
		}

		ctx.Config.Autodetect = partialConfig.Autodetect

		_, _ = reader.Seek(0, io.SeekStart)
		definedProjects = hasLineStartingWith(reader, "projects:")
	}

	var autoProjects []hcl.DetectedProject
	detectionOutput, err := providers.Detect(ctx, &config.Project{Path: wd}, false)
	if err != nil {
		if definedProjects {
			logging.Logger.Debug().Err(err).Msg("could not detect providers")
		} else {
			return fmt.Errorf("could not detect providers %w", err)
		}
	}

	if g.treeFile != "" {
		treeFile, err := os.Create(g.treeFile)
		if err != nil {
			logging.Logger.Warn().Msgf("could not create detected tree file: at %s %s", g.treeFile, err)
		} else {
			_, err = treeFile.WriteString(detectionOutput.Tree)
			if err != nil {
				logging.Logger.Warn().Msgf("could not write detected tree file: %s", err)
			}
			_ = treeFile.Close()
		}
	}

	for _, provider := range detectionOutput.Providers {
		if v, ok := provider.(hcl.DetectedProject); ok {
			autoProjects = append(autoProjects, v)
		}
	}

	if definedProjects {
		m, err := vcs.MetadataFetcher.Get(wd, nil)
		if err != nil {
			logging.Logger.Warn().Msgf("could not fetch git metadata err: %s, default template variables will be blank", err)
		}

		detectedProjects := make([]template.DetectedProject, len(autoProjects))
		detectedPaths := map[string][]template.DetectedProject{}
		for i, p := range autoProjects {
			detectedProjects[i] = template.DetectedProject{
				Name:              p.ProjectName(),
				Path:              p.RelativePath(),
				Env:               p.EnvName(),
				TerraformVarFiles: p.VarFiles(),
				DependencyPaths:   p.DependencyPaths(),
			}

			if v, ok := detectedPaths[p.RelativePath()]; ok {
				detectedPaths[p.RelativePath()] = append(v, detectedProjects[i])
			} else {
				detectedPaths[p.RelativePath()] = []template.DetectedProject{detectedProjects[i]}
			}
		}

		var detectedRootModules []template.DetectedRooModule
		for path, projects := range detectedPaths {
			detectedRootModules = append(detectedRootModules, template.DetectedRooModule{
				Path:     path,
				Projects: projects,
			})
		}

		sort.Slice(detectedRootModules, func(i, j int) bool {
			return detectedRootModules[i].Path < detectedRootModules[j].Path
		})

		variables := template.Variables{
			RepoName:            m.Remote.Name,
			Branch:              m.Branch.Name,
			DetectedProjects:    detectedProjects,
			DetectedRootModules: detectedRootModules,
		}
		if m.PullRequest != nil {
			variables.BaseBranch = m.PullRequest.BaseBranch
		}

		parser := template.NewParser(wd, variables, ctx.Config)
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
	} else {
		buf.WriteString("version: 0.1\n\nprojects:\n")
		for _, p := range autoProjects {
			buf.WriteString(p.YAML())
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

// hasLineStartingWith checks if a file contains a line starting with a specified prefix.
func hasLineStartingWith(reader io.Reader, prefix string) bool {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if strings.HasPrefix(strings.TrimSpace(scanner.Text()), prefix) {
			return true
		}
	}

	return false
}

// unmarshalAutoDetectSection unmarshals the autodetect section of a template
// file. This is required because config templates are not valid YAML and thus
// the native yaml.Unmarshal function cannot be used as it throws an error. To
// get around this we seek the file and read the autodetect section into partial
// config which can be unmarshalled.
func unmarshalAutoDetectSection(reader io.Reader) (*config.Config, error) {
	r := bufio.NewReader(reader)
	var autodetectSection bytes.Buffer
	var recording bool
	var autodetectIndentation int
	for {
		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		if line == "autodetect:\n" {
			recording = true
			autodetectIndentation = getIndentation(line)
		} else if recording && getIndentation(line) <= autodetectIndentation && line != "\n" {
			break
		}

		if recording {
			autodetectSection.WriteString(line)
		}

		if err == io.EOF {
			break
		}
	}

	var partial config.Config
	if err := yaml.Unmarshal(autodetectSection.Bytes(), &partial); err != nil {
		return nil, fmt.Errorf("yaml unmarshal failed: %w", err)
	}

	return &partial, nil
}

func getIndentation(s string) int {
	for i, c := range s {
		if c != ' ' {
			return i
		}
	}
	return len(s)
}
