package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/events"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/prices"
	"github.com/infracost/infracost/internal/providers"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
	"github.com/infracost/infracost/internal/usage"
	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func addRunFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("path", "p", "", "Path to the Terraform directory or JSON/plan file")

	cmd.Flags().String("config-file", "", "Path to Infracost config file. Cannot be used with path, terraform* or usage-file flags")
	cmd.Flags().String("usage-file", "", "Path to Infracost usage file that specifies values for usage-based resources")

	cmd.Flags().String("terraform-plan-flags", "", "Flags to pass to 'terraform plan'. Applicable when path is a Terraform directory")
	cmd.Flags().String("terraform-workspace", "", "Terraform workspace to use. Applicable when path is a Terraform directory")

	cmd.Flags().Bool("show-skipped", false, "Show unsupported resources, some of which might be free")

	cmd.Flags().Bool("sync-usage-file", false, "Sync usage-file with missing resources, needs usage-file too (experimental)")
}

func runMain(cmd *cobra.Command, cfg *config.Config) error {
	projects := make([]*schema.Project, 0)

	for _, projectCfg := range cfg.Projects {
		provider, err := providers.Detect(cfg, projectCfg)

		if err != nil {
			m := fmt.Sprintf("%s\n\n", err)
			m += fmt.Sprintf("Use the %s flag to specify the path to one of the following:\n", ui.PrimaryString("--path"))
			m += " - Terraform plan JSON file\n - Terraform directory\n - Terraform plan file"

			if cmd.Name() != "diff" {
				m += "\n - Terraform state JSON file"
			}

			return events.NewError(errors.New(m), "Could not detect path type")
		}

		if cmd.Name() == "diff" && provider.Type() == "terraform_state_json" {
			m := "Cannot use Terraform state JSON with the infracost diff command.\n\n"
			m += fmt.Sprintf("Use the %s flag to specify the path to one of the following:\n", ui.PrimaryString("--path"))
			m += " - Terraform plan JSON file\n - Terraform directory\n - Terraform plan file"
			return events.NewError(errors.New(m), "Cannot use Terraform state JSON with the infracost diff command")
		}

		m := fmt.Sprintf("Detected %s at %s", provider.DisplayType(), ui.DisplayPath(projectCfg.Path))
		if cfg.IsLogging() {
			log.Info(m)
		} else {
			fmt.Fprintln(os.Stderr, m)
		}

		cfg.Environment.SetProjectEnvironment(provider.Type(), projectCfg)

		u, err := usage.LoadFromFile(projectCfg.UsageFile, cfg.SyncUsageFile)
		if err != nil {
			return err
		}
		if len(u) > 0 {
			cfg.Environment.HasUsageFile = true
		}

		project, err := provider.LoadResources(u)
		if err != nil {
			return err
		}

		projects = append(projects, project)

		if cfg.SyncUsageFile {
			err = usage.SyncUsageData(project, u, projectCfg.UsageFile)
			if err != nil {
				return err
			}
		}

		if !cfg.IsLogging() {
			fmt.Fprintln(os.Stderr, "")
		}
	}

	spinnerOpts := ui.SpinnerOptions{
		EnableLogging: cfg.IsLogging(),
		NoColor:       cfg.NoColor,
	}
	spinner := ui.NewSpinner("Calculating monthly cost estimate", spinnerOpts)

	for _, project := range projects {
		if err := prices.PopulatePrices(cfg, project); err != nil {
			spinner.Fail()
			fmt.Fprintln(os.Stderr, "")

			if e := unwrapped(err); errors.Is(e, prices.ErrInvalidAPIKey) {
				return errors.New(fmt.Sprintf("%v\n%s %s %s %s %s\n%s",
					e.Error(),
					"Please check your",
					ui.PrimaryString(config.CredentialsFilePath()),
					"file or",
					ui.PrimaryString("INFRACOST_API_KEY"),
					"environment variable.",
					"If you continue having issues please email hello@infracost.io",
				))
			}

			if e, ok := err.(*prices.PricingAPIError); ok {
				return errors.New(fmt.Sprintf("%v\n%s", e.Error(), "We have been notified of this issue."))
			}

			return err
		}

		schema.CalculateCosts(project)
		project.CalculateDiff()
	}

	spinner.Success()

	r := output.ToOutputFormat(projects)

	opts := output.Options{
		ShowSkipped: cfg.ShowSkipped,
		NoColor:     cfg.NoColor,
	}

	var (
		b   []byte
		out string
		err error
	)

	switch strings.ToLower(cfg.Format) {
	case "json":
		b, err = output.ToJSON(r, opts)
		out = string(b)
	case "html":
		b, err = output.ToHTML(r, opts)
		out = string(b)
	case "diff":
		b, err = output.ToDiff(r, opts)
		out = fmt.Sprintf("\n%s", string(b))
	case "table_deprecated":
		b, err = output.ToTableDeprecated(r, opts)
		out = fmt.Sprintf("\n%s", string(b))
	default:
		b, err = output.ToTable(r, opts)
		out = fmt.Sprintf("\n%s", string(b))
	}

	if err != nil {
		return errors.Wrap(err, "Error generating output")
	}

	fmt.Printf("%s\n", out)

	return nil
}

func loadRunFlags(cfg *config.Config, cmd *cobra.Command) error {
	hasPathFlag := cmd.Flags().Changed("path")
	hasConfigFile := cmd.Flags().Changed("config-file")

	if cmd.Name() != "infracost" && !hasPathFlag && !hasConfigFile {
		m := fmt.Sprintf("No path specified\n\nUse the %s flag to specify the path to one of the following:\n", ui.PrimaryString("--path"))
		m += " - Terraform plan JSON file\n - Terraform directory\n - Terraform plan file\n - Terraform state JSON file"
		m += "\n\nAlternatively, use --config-file to process multiple projects, see https://infracost.io/config-file"

		ui.PrintUsageErrorAndExit(cmd, m)
	}

	hasProjectFlags := (hasPathFlag ||
		cmd.Flags().Changed("usage-file") ||
		cmd.Flags().Changed("terraform-plan-flags") ||
		cmd.Flags().Changed("terraform-workspace") ||
		cmd.Flags().Changed("terraform-use-state"))

	if hasConfigFile && hasProjectFlags {
		m := "--config-file flag cannot be used with the following flags: "
		m += "--path, --terraform-*, --usage-file"
		ui.PrintUsageErrorAndExit(cmd, m)
	}

	if hasConfigFile {
		cfgFilePath, _ := cmd.Flags().GetString("config-file")
		err := cfg.LoadFromConfigFile(cfgFilePath)

		if err != nil {
			return err
		}
	}

	projectCfg := &config.Project{}

	if hasProjectFlags {
		cfg.Projects = []*config.Project{
			projectCfg,
		}
	}

	if !hasConfigFile {
		err := cfg.LoadFromEnv()
		if err != nil {
			return err
		}
	}

	if hasProjectFlags {
		projectCfg.Path, _ = cmd.Flags().GetString("path")
		projectCfg.UsageFile, _ = cmd.Flags().GetString("usage-file")
		projectCfg.TerraformPlanFlags, _ = cmd.Flags().GetString("terraform-plan-flags")
		projectCfg.TerraformWorkspace, _ = cmd.Flags().GetString("terraform-workspace")
		projectCfg.TerraformUseState, _ = cmd.Flags().GetBool("terraform-use-state")
	}

	cfg.Format, _ = cmd.Flags().GetString("format")
	cfg.ShowSkipped, _ = cmd.Flags().GetBool("show-skipped")
	cfg.SyncUsageFile, _ = cmd.Flags().GetBool("sync-usage-file")

	return nil
}

func checkRunConfig(cfg *config.Config) error {
	if cfg.Format == "json" && cfg.ShowSkipped {
		ui.PrintWarning("show-skipped is not needed with JSON output format as that always includes them.\n")
	}

	if cfg.SyncUsageFile {
		missingUsageFile := make([]string, 0)
		for _, project := range cfg.Projects {
			if project.UsageFile == "" {
				missingUsageFile = append(missingUsageFile, project.Path)
			}
		}
		if len(missingUsageFile) == 1 {
			ui.PrintWarning("Ignoring sync-usage-file as no usage-file is specified.\n")
		} else if len(missingUsageFile) == len(cfg.Projects) {
			ui.PrintWarning("Ignoring sync-usage-file since no projects have a usage-file specified.\n")
		} else if len(missingUsageFile) > 1 {
			ui.PrintWarning(fmt.Sprintf("Ignoring sync-usage-file for following projects as no usage-file is specified for them: %s.\n", strings.Join(missingUsageFile, ", ")))
		}
	}

	return nil
}

func unwrapped(err error) error {
	e := err
	for errors.Unwrap(e) != nil {
		e = errors.Unwrap(e)
	}

	return e
}
