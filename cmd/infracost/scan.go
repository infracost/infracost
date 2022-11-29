package main

import (
	"errors"
	"fmt"

	"github.com/pterm/pterm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/providers/terraform"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
	"github.com/infracost/infracost/internal/usage"
)

type scanCmd struct {
	TerraformVarFiles  []string
	TerraformVars      []string
	TerraformWorkspace string

	Path       string
	ConfigFile string
	UsageFile  string

	cmd *cobra.Command
}

func (s scanCmd) printUsage() {
	ui.PrintUsage(s.cmd)
}
func (s scanCmd) hasProjectFlags() bool {
	return s.Path != "" || len(s.TerraformVars) > 0 || len(s.TerraformVarFiles) > 0 || s.TerraformWorkspace != ""
}

func (s scanCmd) loadRunFlags(cfg *config.Config) error {
	if s.Path == "" && s.ConfigFile == "" {
		m := fmt.Sprintf("No path specified\n\nUse the %s flag to specify the path to one of the following:\n", ui.PrimaryString("--path"))
		m += fmt.Sprintf(" - Terraform/Terragrunt directory\n - Terraform plan JSON file, see %s for how to generate this.", ui.SecondaryLinkString("https://infracost.io/troubleshoot"))
		m += fmt.Sprintf("\n\nAlternatively, use --config-file to process multiple projects, see %s", ui.SecondaryLinkString("https://infracost.io/config-file"))

		s.printUsage()
		return errors.New(m)
	}

	if s.ConfigFile != "" && s.hasProjectFlags() {
		m := "--config-file flag cannot be used with the following flags: "
		m += "--path, --project-name, --terraform-*, --usage-file"
		s.printUsage()
		return errors.New(m)
	}

	projectCfg := cfg.Projects[0]

	if s.hasProjectFlags() {
		cfg.RootPath = s.Path
		projectCfg.Path = s.Path

		projectCfg.TerraformVarFiles = s.TerraformVarFiles
		projectCfg.TerraformVars = tfVarsToMap(s.TerraformVars)
		projectCfg.UsageFile = s.UsageFile
		projectCfg.TerraformWorkspace = s.TerraformWorkspace
	}

	if s.ConfigFile != "" {
		err := cfg.LoadFromConfigFile(s.ConfigFile)
		if err != nil {
			return err
		}

		cfg.ConfigFilePath = s.ConfigFile
	}

	return nil
}

func (s scanCmd) run(runCtx *config.RunContext) error {
	err := s.loadRunFlags(runCtx.Config)
	if err != nil {
		return err
	}

	spinnerOpts := ui.SpinnerOptions{
		EnableLogging: runCtx.Config.IsLogging(),
		NoColor:       runCtx.Config.NoColor,
		Indent:        "  ",
	}
	spinner := ui.NewSpinner("Scanning projects for cost optimizations...", spinnerOpts)
	defer spinner.Fail()

	usageData := make(map[string]*schema.UsageData)

	if s.UsageFile != "" {
		usageFile, err := usage.LoadUsageFile(s.UsageFile)
		if err != nil {
			return err
		}

		usageData = usageFile.ToUsageDataMap()
	}

	var projects []*schema.Project
	for _, project := range runCtx.Config.Projects {
		projectCtx := config.NewProjectContext(runCtx, project, log.Fields{})
		hclProvider, err := terraform.NewHCLProvider(projectCtx, &terraform.HCLProviderConfig{SuppressLogging: true})
		if err != nil {
			logging.Logger.WithError(err).Errorf("failed to load a provider for path %s", project.Path)
			continue
		}

		hclProjects, err := hclProvider.LoadResources(usageData)
		if err != nil {
			logging.Logger.WithError(err).Errorf("failed to load Terraform projects for path %s", project.Path)
			continue
		}

		projects = append(projects, hclProjects...)
	}

	spinner.Success()

	for _, project := range projects {
		if len(project.Metadata.Policies) == 0 {
			pterm.DefaultBox.WithTitle(project.Name).Println(pterm.Green("No cost optimizations found."))
			continue
		}

		rows := make([][]string, len(project.Metadata.Policies)+1)
		rows[0] = []string{"Address", "Title", "Cost Saving"}
		for i, suggestion := range project.Metadata.Policies {
			cost := "?"
			if suggestion.Cost != nil {
				cost = output.FormatCost2DP(runCtx.Config.Currency, suggestion.Cost)
			}

			rows[i+1] = []string{suggestion.Address, suggestion.Title, cost}
		}

		pterm.DefaultBox.WithTitle(project.Name).Println(pterm.DefaultTable.WithHasHeader().WithData(rows).Srender())
	}

	return nil
}

func scanCommand(ctx *config.RunContext) *cobra.Command {
	var scan scanCmd

	cmd := &cobra.Command{
		Use:    "scan",
		Short:  "Scan your Terraform projects for cost optimizations",
		Long:   "Scan your Terraform projects for cost optimizations",
		Hidden: true,
		Example: `
      infracost scan --path /code --terraform-var-file my.tfvars
      `,
		ValidArgs: []string{"--", "-"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkAPIKey(ctx.Config.APIKey, ctx.Config.PricingAPIEndpoint, ctx.Config.DefaultPricingAPIEndpoint); err != nil {
				return err
			}

			return scan.run(ctx)
		},
	}

	scan.cmd = cmd
	cmd.Flags().StringSliceVar(&scan.TerraformVarFiles, "terraform-var-file", nil, "Load variable files, similar to Terraform's -var-file flag. Provided files must be relative to the --path flag")
	cmd.Flags().StringSliceVar(&scan.TerraformVars, "terraform-var", nil, "Set value for an input variable, similar to Terraform's -var flag")
	cmd.Flags().StringVar(&scan.TerraformWorkspace, "terraform-workspace", "", "Terraform workspace to use. Applicable when path is a Terraform directory")

	cmd.Flags().StringVarP(&scan.Path, "path", "p", "", "Path to the Terraform directory or JSON/plan file")
	cmd.Flags().StringVar(&scan.ConfigFile, "config-file", "", "Path to Infracost config file. Cannot be used with path, terraform* or usage-file flags")
	cmd.Flags().StringVar(&scan.UsageFile, "usage-file", "", "Path to Infracost usage file that specifies values for usage-based resources")

	return cmd
}
