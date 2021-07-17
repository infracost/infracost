package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
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

	_ = cmd.MarkFlagFilename("path", "json", "tf")
	_ = cmd.MarkFlagFilename("config-file", "yml")
	_ = cmd.MarkFlagFilename("usage-file", "yml")
}

func runMain(cmd *cobra.Command, runCtx *config.RunContext) error {
	projects := make([]*schema.Project, 0)
	projectContexts := make([]*config.ProjectContext, 0)

	for _, projectCfg := range runCtx.Config.Projects {
		ctx := config.NewProjectContext(runCtx, projectCfg)
		runCtx.SetCurrentProjectContext(ctx)

		provider, err := providers.Detect(ctx)
		if err != nil {
			m := fmt.Sprintf("%s\n\n", err)
			m += fmt.Sprintf("Use the %s flag to specify the path to one of the following:\n", ui.PrimaryString("--path"))
			m += " - Terraform plan JSON file\n - Terraform directory\n - Terraform plan file"

			if cmd.Name() != "diff" {
				m += "\n - Terraform state JSON file"
			}

			return clierror.NewSanitizedError(errors.New(m), "Could not detect path type")
		}
		ctx.SetContextValue("projectType", provider.Type())
		projectContexts = append(projectContexts, ctx)

		if cmd.Name() == "diff" && provider.Type() == "terraform_state_json" {
			m := "Cannot use Terraform state JSON with the infracost diff command.\n\n"
			m += fmt.Sprintf("Use the %s flag to specify the path to one of the following:\n", ui.PrimaryString("--path"))
			m += " - Terraform plan JSON file\n - Terraform directory\n - Terraform plan file"
			return clierror.NewSanitizedError(errors.New(m), "Cannot use Terraform state JSON with the infracost diff command")
		}

		m := fmt.Sprintf("Detected %s at %s", provider.DisplayType(), ui.DisplayPath(projectCfg.Path))
		if runCtx.Config.IsLogging() {
			log.Info(m)
		} else {
			fmt.Fprintln(os.Stderr, m)
		}

		u, err := usage.LoadFromFile(projectCfg.UsageFile, runCtx.Config.SyncUsageFile)
		if err != nil {
			return err
		}
		if len(u) > 0 {
			ctx.SetContextValue("hasUsageFile", true)
		}

		metadata := config.DetectProjectMetadata(ctx)
		metadata.Type = provider.Type()
		provider.AddMetadata(metadata)
		name := schema.GenerateProjectName(metadata, runCtx.Config.EnableDashboard)

		project := schema.NewProject(name, metadata)
		err = provider.LoadResources(project, u)
		if err != nil {
			return err
		}

		projects = append(projects, project)

		if runCtx.Config.SyncUsageFile {
			err = usage.SyncUsageData(project, u, projectCfg.UsageFile)
			if err != nil {
				return err
			}
		}

		if !runCtx.Config.IsLogging() {
			fmt.Fprintln(os.Stderr, "")
		}
	}

	spinnerOpts := ui.SpinnerOptions{
		EnableLogging: runCtx.Config.IsLogging(),
		NoColor:       runCtx.Config.NoColor,
	}
	spinner := ui.NewSpinner("Calculating monthly cost estimate", spinnerOpts)

	for _, project := range projects {
		if err := prices.PopulatePrices(runCtx.Config, project); err != nil {
			spinner.Fail()
			fmt.Fprintln(os.Stderr, "")

			if e := unwrapped(err); errors.Is(e, apiclient.ErrInvalidAPIKey) {
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

			if e, ok := err.(*apiclient.APIError); ok {
				return errors.New(fmt.Sprintf("%v\n%s", e.Error(), "We have been notified of this issue."))
			}

			return err
		}

		schema.CalculateCosts(project)
		project.CalculateDiff()
	}

	spinner.Success()

	r := output.ToOutputFormat(projects)

	var err error

	c := apiclient.NewDashboardAPIClient(runCtx)
	r.RunID, err = c.AddRun(runCtx, projectContexts, r)
	if err != nil {
		log.Errorf("Error reporting run: %s", err)
	}

	env := buildRunEnv(runCtx, projectContexts, r)

	err = c.AddEvent("infracost-run", env)
	if err != nil {
		log.Errorf("Error reporting event: %s", err)
	}

	opts := output.Options{
		DashboardEnabled: runCtx.Config.EnableDashboard,
		ShowSkipped:      runCtx.Config.ShowSkipped,
		NoColor:          runCtx.Config.NoColor,
		Fields:           runCtx.Config.Fields,
	}

	var (
		b   []byte
		out string
	)

	switch strings.ToLower(runCtx.Config.Format) {
	case "json":
		b, err = output.ToJSON(r, opts)
		out = string(b)
	case "html":
		b, err = output.ToHTML(r, opts)
		out = string(b)
	case "diff":
		b, err = output.ToDiff(r, opts)
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

	validFields := []string{"price", "monthlyQuantity", "unit", "hourlyCost", "monthlyCost"}
	validFieldsFormats := []string{"table", "html"}

	if cmd.Flags().Changed("fields") {
		if c, _ := cmd.Flags().GetStringSlice("fields"); len(c) == 0 {
			ui.PrintWarningf("fields is empty, using defaults: %s", cmd.Flag("fields").DefValue)
		} else if cfg.Fields != nil && !contains(validFieldsFormats, cfg.Format) {
			ui.PrintWarning("fields is only supported for table and html output formats")
		} else {
			fields, _ := cmd.Flags().GetStringSlice("fields")
			vf := []string{}
			for _, f := range fields {
				if !contains(validFields, f) {
					ui.PrintWarningf("Invalid field '%s' specified, valid fields are: %s", f, validFields)
				} else {
					vf = append(vf, f)
				}
			}
			cfg.Fields = vf
		}
	}

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

func buildRunEnv(runCtx *config.RunContext, projectContexts []*config.ProjectContext, r output.Root) map[string]interface{} {
	env := runCtx.EventEnvWithProjectContexts(projectContexts)
	env["projectCount"] = len(projectContexts)

	summary := r.FullSummary
	env["supportedResourceCounts"] = summary.SupportedResourceCounts
	env["unsupportedResourceCounts"] = summary.UnsupportedResourceCounts
	env["totalSupportedResources"] = summary.TotalSupportedResources
	env["totalUnsupportedResources"] = summary.TotalUnsupportedResources
	env["totalNoPriceResources"] = summary.TotalNoPriceResources
	env["totalResources"] = summary.TotalResources

	return env
}

func unwrapped(err error) error {
	e := err
	for errors.Unwrap(e) != nil {
		e = errors.Unwrap(e)
	}

	return e
}
