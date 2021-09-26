package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Rhymond/go-money"

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

	var spinner *ui.Spinner

	for _, projectCfg := range runCtx.Config.Projects {
		ctx := config.NewProjectContext(runCtx, projectCfg)
		runCtx.SetCurrentProjectContext(ctx)

		provider, err := providers.Detect(ctx)
		if err != nil {
			m := fmt.Sprintf("%s\n\n", err)
			m += fmt.Sprintf("Use the %s flag to specify the path to one of the following:\n", ui.PrimaryString("--path"))
			m += " - Terraform plan JSON file\n - Terraform/Terragrunt directory\n - Terraform plan file"

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
			m += " - Terraform plan JSON file\n - Terraform/Terragrunt directory\n - Terraform plan file"
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

		providerProjects, err := provider.LoadResources(u)
		if err != nil {
			return err
		}

		if runCtx.Config.SyncUsageFile {
			spinnerOpts := ui.SpinnerOptions{
				EnableLogging: runCtx.Config.IsLogging(),
				NoColor:       runCtx.Config.NoColor,
				Indent:        "  ",
			}
			spinner = ui.NewSpinner("Syncing usage data from cloud", spinnerOpts)

			syncResult, err := usage.SyncUsageData(providerProjects, u, projectCfg.UsageFile)
			summarizeUsage(ctx, syncResult)
			if err != nil {
				spinner.Fail()
				return err
			}

			remediateUsage(runCtx, ctx, syncResult)

			u, err := usage.LoadFromFile(projectCfg.UsageFile, runCtx.Config.SyncUsageFile)
			if err != nil {
				spinner.Fail()
				return err
			}
			providerProjects, err = provider.LoadResources(u)
			if err != nil {
				spinner.Fail()
				return err
			}

			resources := syncResult.ResourceCount
			attempts := syncResult.EstimationCount
			errors := len(syncResult.EstimationErrors)
			successes := attempts - errors

			pluralized := ""
			if resources > 1 {
				pluralized = "s"
			}

			spinner.Success()
			cmd.Println(fmt.Sprintf("    %s Synced %d of %d resource%s",
				ui.FaintString("└─"),
				successes,
				resources,
				pluralized))
		}

		projects = append(projects, providerProjects...)

		if !runCtx.Config.IsLogging() {
			fmt.Fprintln(os.Stderr, "")
		}
	}

	spinnerOpts := ui.SpinnerOptions{
		EnableLogging: runCtx.Config.IsLogging(),
		NoColor:       runCtx.Config.NoColor,
	}
	spinner = ui.NewSpinner("Calculating monthly cost estimate", spinnerOpts)

	for _, project := range projects {
		if err := prices.PopulatePrices(runCtx.Config, project); err != nil {
			spinner.Fail()
			fmt.Fprintln(os.Stderr, "")

			if e := unwrapped(err); errors.Is(e, apiclient.ErrInvalidAPIKey) {
				return fmt.Errorf("%v\n%s %s %s %s %s\n%s",
					e.Error(),
					"Please check your",
					ui.PrimaryString(config.CredentialsFilePath()),
					"file or",
					ui.PrimaryString("INFRACOST_API_KEY"),
					"environment variable.",
					"If you continue having issues please email hello@infracost.io",
				)
			}

			if e, ok := err.(*apiclient.APIError); ok {
				return fmt.Errorf("%v\n%s", e.Error(), "We have been notified of this issue.")
			}

			return err
		}

		schema.CalculateCosts(project)
		project.CalculateDiff()
	}

	spinner.Success()

	r := output.ToOutputFormat(projects)
	r.Currency = runCtx.Config.Currency

	var err error

	dashboardClient := apiclient.NewDashboardAPIClient(runCtx)
	r.RunID, err = dashboardClient.AddRun(runCtx, projectContexts, r)
	if err != nil {
		log.Errorf("Error reporting run: %s", err)
	}

	env := buildRunEnv(runCtx, projectContexts, r)

	pricingClient := apiclient.NewPricingAPIClient(runCtx.Config)
	err = pricingClient.AddEvent("infracost-run", env)
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

	cmd.Printf("%s\n", out)

	return nil
}

func summarizeUsage(ctx *config.ProjectContext, syncResult *usage.SyncResult) {
	var usageSyncs, usageEstimates, usageEstimateErrors int
	if syncResult != nil {
		usageSyncs = syncResult.ResourceCount
		usageEstimates = syncResult.EstimationCount
		usageEstimateErrors = len(syncResult.EstimationErrors)
	}
	ctx.SetContextValue("usageSyncs", usageSyncs)
	ctx.SetContextValue("usageEstimates", usageEstimates)
	ctx.SetContextValue("usageEstimateErrors", usageEstimateErrors)
}

func remediateUsage(runCtx *config.RunContext, ctx *config.ProjectContext, syncResult *usage.SyncResult) {
	var remAttempts, remErrors int
	for name, err := range syncResult.EstimationErrors {
		if remediater, ok := err.(schema.Remediater); ok {
			remAttempts++
			err = remediater.Remediate()
			if err != nil {
				remErrors++
				log.Warningf("Cannot enable estimation for %s: %s", name, err.Error())
			}
		}
	}
	ctx.SetContextValue("remediationAttempts", remAttempts)
	ctx.SetContextValue("remediationErrors", remErrors)
}

func loadRunFlags(cfg *config.Config, cmd *cobra.Command) error {
	hasPathFlag := cmd.Flags().Changed("path")
	hasConfigFile := cmd.Flags().Changed("config-file")

	if cmd.Name() != "infracost" && !hasPathFlag && !hasConfigFile {
		m := fmt.Sprintf("No path specified\n\nUse the %s flag to specify the path to one of the following:\n", ui.PrimaryString("--path"))
		m += " - Terraform plan JSON file\n - Terraform/Terragrunt directory\n - Terraform plan file\n - Terraform state JSON file"
		m += "\n\nAlternatively, use --config-file to process multiple projects, see https://infracost.io/config-file"

		ui.PrintUsage(cmd)
		return errors.New(m)
	}

	hasProjectFlags := (hasPathFlag ||
		cmd.Flags().Changed("usage-file") ||
		cmd.Flags().Changed("terraform-plan-flags") ||
		cmd.Flags().Changed("terraform-workspace") ||
		cmd.Flags().Changed("terraform-use-state"))

	projectCfg := cfg.Projects[0]

	hasProjectEnvs := projectCfg.Path != "" ||
		projectCfg.TerraformBinary != "" ||
		projectCfg.TerraformCloudHost != "" ||
		projectCfg.TerraformWorkspace != "" ||
		projectCfg.TerraformCloudToken != ""

	if hasConfigFile && (hasProjectFlags || hasProjectEnvs) {
		m := "--config-file flag cannot be used with the following flags or equivalent environment variables: "
		m += "--path, --terraform-*, --usage-file"
		ui.PrintUsage(cmd)
		return errors.New(m)
	}

	if hasProjectFlags {
		projectCfg.Path, _ = cmd.Flags().GetString("path")
		projectCfg.UsageFile, _ = cmd.Flags().GetString("usage-file")
		projectCfg.TerraformPlanFlags, _ = cmd.Flags().GetString("terraform-plan-flags")
		projectCfg.TerraformUseState, _ = cmd.Flags().GetBool("terraform-use-state")

		if cmd.Flags().Changed("terraform-workspace") {
			projectCfg.TerraformWorkspace, _ = cmd.Flags().GetString("terraform-workspace")
		}
	}

	if hasConfigFile {
		cfgFilePath, _ := cmd.Flags().GetString("config-file")
		err := cfg.LoadFromConfigFile(cfgFilePath)

		if err != nil {
			return err
		}
	}

	cfg.Format, _ = cmd.Flags().GetString("format")
	cfg.ShowSkipped, _ = cmd.Flags().GetBool("show-skipped")
	cfg.SyncUsageFile, _ = cmd.Flags().GetBool("sync-usage-file")

	includeAllFields := "all"
	validFields := []string{"price", "monthlyQuantity", "unit", "hourlyCost", "monthlyCost"}
	validFieldsFormats := []string{"table", "html"}

	if cmd.Flags().Changed("fields") {
		fields, _ := cmd.Flags().GetStringSlice("fields")
		if len(fields) == 0 {
			ui.PrintWarningf(cmd.ErrOrStderr(), "fields is empty, using defaults: %s", cmd.Flag("fields").DefValue)
		} else if cfg.Fields != nil && !contains(validFieldsFormats, cfg.Format) {
			ui.PrintWarning(cmd.ErrOrStderr(), "fields is only supported for table and html output formats")
		} else if len(fields) == 1 && fields[0] == includeAllFields {
			cfg.Fields = validFields
		} else {
			vf := []string{}
			for _, f := range fields {
				if !contains(validFields, f) {
					ui.PrintWarningf(cmd.ErrOrStderr(), "Invalid field '%s' specified, valid fields are: %s or '%s' to include all fields", f, validFields, includeAllFields)
				} else {
					vf = append(vf, f)
				}
			}
			cfg.Fields = vf
		}
	}

	return nil
}

func checkRunConfig(warningWriter io.Writer, cfg *config.Config) error {
	if cfg.Format == "json" && cfg.ShowSkipped {
		ui.PrintWarning(warningWriter, "show-skipped is not needed with JSON output format as that always includes them.\n")
	}

	if cfg.SyncUsageFile {
		missingUsageFile := make([]string, 0)
		for _, project := range cfg.Projects {
			if project.UsageFile == "" {
				missingUsageFile = append(missingUsageFile, project.Path)
			}
		}
		if len(missingUsageFile) == 1 {
			ui.PrintWarning(warningWriter, "Ignoring sync-usage-file as no usage-file is specified.\n")
		} else if len(missingUsageFile) == len(cfg.Projects) {
			ui.PrintWarning(warningWriter, "Ignoring sync-usage-file since no projects have a usage-file specified.\n")
		} else if len(missingUsageFile) > 1 {
			ui.PrintWarning(warningWriter, fmt.Sprintf("Ignoring sync-usage-file for following projects as no usage-file is specified for them: %s.\n", strings.Join(missingUsageFile, ", ")))
		}
	}

	if money.GetCurrency(cfg.Currency) == nil {
		ui.PrintWarning(warningWriter, fmt.Sprintf("Ignoring unknown currency '%s', using USD.\n", cfg.Currency))
		cfg.Currency = "USD"
	}

	return nil
}

func buildRunEnv(runCtx *config.RunContext, projectContexts []*config.ProjectContext, r output.Root) map[string]interface{} {
	env := runCtx.EventEnvWithProjectContexts(projectContexts)
	env["projectCount"] = len(projectContexts)
	env["runSeconds"] = time.Now().Unix() - runCtx.StartTime
	env["currency"] = runCtx.Config.Currency

	summary := r.FullSummary
	env["supportedResourceCounts"] = summary.SupportedResourceCounts
	env["unsupportedResourceCounts"] = summary.UnsupportedResourceCounts
	env["totalSupportedResources"] = summary.TotalSupportedResources
	env["totalUnsupportedResources"] = summary.TotalUnsupportedResources
	env["totalNoPriceResources"] = summary.TotalNoPriceResources
	env["totalResources"] = summary.TotalResources

	env["estimatedUsageCounts"] = summary.EstimatedUsageCounts
	env["unestimatedUsageCounts"] = summary.UnestimatedUsageCounts
	env["totalEstimatedUsages"] = summary.TotalEstimatedUsages
	env["totalUnestimatedUsages"] = summary.TotalUnestimatedUsages

	return env
}

func unwrapped(err error) error {
	e := err
	for errors.Unwrap(e) != nil {
		e = errors.Unwrap(e)
	}

	return e
}
