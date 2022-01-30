package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Rhymond/go-money"
	"golang.org/x/sync/errgroup"

	"github.com/pkg/errors"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/prices"
	"github.com/infracost/infracost/internal/providers"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
	"github.com/infracost/infracost/internal/usage"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type projectJob = struct {
	index      int
	projectCfg *config.Project
}

type projectResult = struct {
	index    int
	ctx      *config.ProjectContext
	projects []*schema.Project
}

var validRunFormats = []string{"json", "table", "html"}

func addRunFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("path", "p", "", "Path to the Terraform directory or JSON/plan file")

	cmd.Flags().String("config-file", "", "Path to Infracost config file. Cannot be used with path, terraform* or usage-file flags")
	cmd.Flags().String("usage-file", "", "Path to Infracost usage file that specifies values for usage-based resources")

	cmd.Flags().String("terraform-plan-flags", "", "Flags to pass to 'terraform plan'. Applicable when path is a Terraform directory")
	cmd.Flags().String("terraform-workspace", "", "Terraform workspace to use. Applicable when path is a Terraform directory")

	cmd.Flags().Bool("no-cache", false, "Don't attempt to cache Terraform plans")

	cmd.Flags().Bool("show-skipped", false, "Show unsupported resources")

	cmd.Flags().Bool("sync-usage-file", false, "Sync usage-file with missing resources, needs usage-file too (experimental)")

	_ = cmd.MarkFlagFilename("path", "json", "tf")
	_ = cmd.MarkFlagFilename("config-file", "yml")
	_ = cmd.MarkFlagFilename("usage-file", "yml")
}

func runMain(cmd *cobra.Command, runCtx *config.RunContext) error {
	if runCtx.Config.IsSelfHosted() && runCtx.Config.EnableDashboard {
		ui.PrintWarning(cmd.ErrOrStderr(), "The dashboard is part of Infracost's hosted services. Contact hello@infracost.io for help.")
	}

	parallelism, err := getParallelism(cmd, runCtx)
	if err != nil {
		return err
	}
	runCtx.SetContextValue("parallelism", parallelism)

	numJobs := len(runCtx.Config.Projects)
	jobs := make(chan projectJob, numJobs)

	projectResultChan := make(chan projectResult, numJobs)
	errGroup, _ := errgroup.WithContext(context.Background())

	if parallelism > 1 && numJobs > 1 && !runCtx.Config.IsLogging() {
		cmd.PrintErrln("Running multiple projects in parallel, so log-level=info is enabled by default.")
		cmd.PrintErrln("Run with INFRACOST_PARALLELISM=1 to disable parallelism to help debugging.")
		cmd.PrintErrln()

		runCtx.Config.LogLevel = "info"
		err := runCtx.Config.ConfigureLogger()
		if err != nil {
			return err
		}
	}

	// Create a mutex for each path, so we can synchronize the runs of any
	// projects that have the same path. This is necessary because Terraform
	// can't run multiple operations in parallel on the same path.
	pathMuxs := map[string]*sync.Mutex{}
	for _, projectCfg := range runCtx.Config.Projects {
		pathMuxs[projectCfg.Path] = &sync.Mutex{}
	}

	for i := 0; i < parallelism; i++ {
		errGroup.Go(func() error {

			for job := range jobs {
				mux := pathMuxs[job.projectCfg.Path]

				ctx := config.NewProjectContext(runCtx, job.projectCfg)
				configProjects, err := runProjectConfig(cmd, runCtx, ctx, job.projectCfg, mux)
				if err != nil {
					return err
				}

				projectResultChan <- projectResult{
					index:    job.index,
					ctx:      ctx,
					projects: configProjects,
				}
			}

			return nil
		})
	}

	for i, p := range runCtx.Config.Projects {
		jobs <- projectJob{index: i, projectCfg: p}
	}
	close(jobs)

	err = errGroup.Wait()
	if err != nil {
		return err
	}

	close(projectResultChan)
	projectResults := make([]projectResult, 0, len(runCtx.Config.Projects))
	for result := range projectResultChan {
		projectResults = append(projectResults, result)
	}
	sort.Slice(projectResults, func(i, j int) bool {
		return projectResults[i].index < projectResults[j].index
	})

	projects := make([]*schema.Project, 0)
	projectContexts := make([]*config.ProjectContext, 0)
	for _, projectResult := range projectResults {
		for _, project := range projectResult.projects {
			projectContexts = append(projectContexts, projectResult.ctx)
			projects = append(projects, project)
		}
	}

	r, err := output.ToOutputFormat(projects)
	if err != nil {
		return err
	}

	r.IsCIRun = runCtx.IsCIRun()
	r.Currency = runCtx.Config.Currency

	dashboardClient := apiclient.NewDashboardAPIClient(runCtx)
	result, err := dashboardClient.AddRun(runCtx, projectContexts, r)
	if err != nil {
		log.Errorf("Error reporting run: %s", err)
	}

	r.RunID, r.ShareURL = result.RunID, result.ShareURL

	opts := output.Options{
		DashboardEnabled: runCtx.Config.EnableDashboard,
		ShowSkipped:      runCtx.Config.ShowSkipped,
		NoColor:          runCtx.Config.NoColor,
		Fields:           runCtx.Config.Fields,
	}

	var b []byte

	switch strings.ToLower(runCtx.Config.Format) {
	case "json":
		b, err = output.ToJSON(r, opts)
	case "html":
		b, err = output.ToHTML(r, opts)
	case "diff":
		b, err = output.ToDiff(r, opts)
	default:
		b, err = output.ToTable(r, opts)
	}

	if err != nil {
		return errors.Wrap(err, "Error generating output")
	}

	if runCtx.Config.Format == "diff" || runCtx.Config.Format == "table" {
		lines := bytes.Count(b, []byte("\n")) + 1
		runCtx.SetContextValue("lineCount", lines)
	}

	env := buildRunEnv(runCtx, projectContexts, r)
	pricingClient := apiclient.NewPricingAPIClient(runCtx)
	err = pricingClient.AddEvent("infracost-run", env)
	if err != nil {
		log.Errorf("Error reporting event: %s", err)
	}

	// Print a new line to separate the logs from the output
	if runCtx.Config.IsLogging() {
		cmd.PrintErrln()
	}

	if outFile, _ := cmd.Flags().GetString("out-file"); outFile != "" {
		err = saveOutFile(cmd, outFile, b)
		if err != nil {
			return err
		}
	} else {
		cmd.Println(string(b))
	}

	return nil
}

func runProjectConfig(cmd *cobra.Command, runCtx *config.RunContext, ctx *config.ProjectContext, projectCfg *config.Project, mux *sync.Mutex) ([]*schema.Project, error) {
	if mux != nil {
		mux.Lock()
		defer mux.Unlock()
	}

	for k, v := range projectCfg.Env {
		os.Setenv(k, v)
	}

	provider, err := providers.Detect(ctx)
	if err != nil {
		m := fmt.Sprintf("%s\n\n", err)
		m += fmt.Sprintf("Try setting --path to a Terraform plan JSON file. See %s for how to generate this.", ui.LinkString("https://infracost.io/troubleshoot"))

		return []*schema.Project{}, clierror.NewSanitizedError(errors.New(m), "Could not detect path type")
	}
	ctx.SetContextValue("projectType", provider.Type())

	if cmd.Name() == "diff" && provider.Type() == "terraform_state_json" {
		m := "Cannot use Terraform state JSON with the infracost diff command.\n\n"
		m += fmt.Sprintf("Use the %s flag to specify the path to one of the following:\n", ui.PrimaryString("--path"))
		m += " - Terraform plan JSON file\n - Terraform/Terragrunt directory\n - Terraform plan file"
		return []*schema.Project{}, clierror.NewSanitizedError(errors.New(m), "Cannot use Terraform state JSON with the infracost diff command")
	}

	m := fmt.Sprintf("Detected %s at %s", provider.DisplayType(), ui.DisplayPath(projectCfg.Path))
	if runCtx.Config.IsLogging() {
		log.Info(m)
	} else {
		fmt.Fprintln(os.Stderr, m)
	}

	// Generate usage file
	if runCtx.Config.SyncUsageFile {
		err := generateUsageFile(cmd, runCtx, ctx, projectCfg, provider)
		if err != nil {
			return []*schema.Project{}, errors.Wrap(err, "Error generating usage file")
		}
	}

	// Load usage data
	usageData := make(map[string]*schema.UsageData)
	var usageFile *usage.UsageFile

	if projectCfg.UsageFile != "" {
		var err error
		usageFile, err = usage.LoadUsageFile(projectCfg.UsageFile)
		if err != nil {
			return []*schema.Project{}, err
		}

		invalidKeys, err := usageFile.InvalidKeys()
		if err != nil {
			log.Errorf("Error checking usage file keys: %v", err)
		} else if len(invalidKeys) > 0 {
			ui.PrintWarningf(cmd.ErrOrStderr(),
				"The following usage file parameters are invalid and will be ignored: %s\n",
				strings.Join(invalidKeys, ", "),
			)
		}
	} else {
		usageFile = usage.NewBlankUsageFile()
	}

	if len(usageData) > 0 {
		ctx.SetContextValue("hasUsageFile", true)
	}

	// Merge wildcard usages into individual usage
	wildCardUsage := make(map[string]*usage.ResourceUsage)
	for _, us := range usageFile.ResourceUsages {
		if strings.HasSuffix(us.Name, "[*]") {
			lastIndexOfOpenBracket := strings.LastIndex(us.Name, "[")
			prefixName := us.Name[:lastIndexOfOpenBracket]
			wildCardUsage[prefixName] = us
		}
	}

	for _, us := range usageFile.ResourceUsages {
		if strings.HasSuffix(us.Name, "[*]") {
			continue
		}

		if !strings.HasSuffix(us.Name, "]") {
			continue
		}
		lastIndexOfOpenBracket := strings.LastIndex(us.Name, "[")
		prefixName := us.Name[:lastIndexOfOpenBracket]

		us.MergeResourceUsage(wildCardUsage[prefixName])
	}

	usageData = usageFile.ToUsageDataMap()

	projects, err := provider.LoadResources(usageData)
	if err != nil {
		cmd.PrintErrln()
		return projects, err
	}

	spinnerOpts := ui.SpinnerOptions{
		EnableLogging: runCtx.Config.IsLogging(),
		NoColor:       runCtx.Config.NoColor,
		Indent:        "  ",
	}
	spinner := ui.NewSpinner("Retrieving cloud prices to calculate costs", spinnerOpts)
	defer spinner.Fail()

	for _, project := range projects {
		if err := prices.PopulatePrices(runCtx, project); err != nil {
			spinner.Fail()
			cmd.PrintErrln()

			if e := unwrapped(err); errors.Is(e, apiclient.ErrInvalidAPIKey) {
				return projects, fmt.Errorf("%v\n%s %s %s %s %s\n%s",
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
				return projects, fmt.Errorf("%v\n%s", e.Error(), "We have been notified of this issue.")
			}

			return projects, err
		}

		schema.CalculateCosts(project)
		project.CalculateDiff()
	}

	spinner.Success()

	if !runCtx.Config.IsLogging() {
		cmd.PrintErrln()
	}

	return projects, nil
}

func generateUsageFile(cmd *cobra.Command, runCtx *config.RunContext, projectCtx *config.ProjectContext, projectCfg *config.Project, provider schema.Provider) error {
	if projectCfg.UsageFile == "" {
		// This should not happen as we check earlier in the code that usage-file is not empty when sync-usage-file flag is on.
		return fmt.Errorf("Error generating usage: no usage file given")
	}

	var usageFile *usage.UsageFile

	usageFilePath := projectCfg.UsageFile
	err := usage.CreateUsageFile(usageFilePath)
	if err != nil {
		return errors.Wrap(err, "Error creating usage file")
	}

	usageFile, err = usage.LoadUsageFile(usageFilePath)
	if err != nil {
		return errors.Wrap(err, "Error loading usage file")
	}

	usageData := usageFile.ToUsageDataMap()
	providerProjects, err := provider.LoadResources(usageData)
	if err != nil {
		return errors.Wrap(err, "Error loading resources")
	}

	spinnerOpts := ui.SpinnerOptions{
		EnableLogging: runCtx.Config.IsLogging(),
		NoColor:       runCtx.Config.NoColor,
		Indent:        "  ",
	}

	spinner := ui.NewSpinner("Syncing usage data from cloud", spinnerOpts)
	defer spinner.Fail()

	syncResult, err := usage.SyncUsageData(usageFile, providerProjects)

	if err != nil {
		spinner.Fail()
		return errors.Wrap(err, "Error synchronizing usage data")
	}

	projectCtx.SetFrom(syncResult)
	if err != nil {
		spinner.Fail()
		return errors.Wrap(err, "Error summarizing usage")
	}

	err = usageFile.WriteToPath(projectCfg.UsageFile)
	if err != nil {
		spinner.Fail()
		return errors.Wrap(err, "Error writing usage file")
	}

	if syncResult == nil {
		spinner.Fail()
	} else {
		resources := syncResult.ResourceCount
		attempts := syncResult.EstimationCount
		errors := len(syncResult.EstimationErrors)
		successes := attempts - errors

		pluralized := ""
		if resources > 1 {
			pluralized = "s"
		}

		spinner.Success()
		cmd.PrintErrln(fmt.Sprintf("    %s Synced %d of %d resource%s",
			ui.FaintString("└─"),
			successes,
			resources,
			pluralized))
	}
	return nil
}

func getParallelism(cmd *cobra.Command, runCtx *config.RunContext) (int, error) {
	var parallelism int

	if runCtx.Config.Parallelism == nil {
		parallelism = 4
		numCPU := runtime.NumCPU()
		if numCPU*4 > parallelism {
			parallelism = numCPU * 4
		}
		if parallelism > 16 {
			parallelism = 16
		}
	} else {
		parallelism = *runCtx.Config.Parallelism

		if parallelism < 0 {
			return parallelism, fmt.Errorf("parallelism must be a positive number")
		}

		if parallelism > 16 {
			return parallelism, fmt.Errorf("parallelism must be less than 16")
		}
	}

	return parallelism, nil
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

	if hasConfigFile && hasProjectFlags {
		m := "--config-file flag cannot be used with the following flags: "
		m += "--path, --terraform-*, --usage-file"
		ui.PrintUsage(cmd)
		return errors.New(m)
	}

	projectCfg := cfg.Projects[0]

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

	cfg.NoCache, _ = cmd.Flags().GetBool("no-cache")

	cfg.Format, _ = cmd.Flags().GetString("format")

	if cfg.Format != "" && !contains(validRunFormats, cfg.Format) {
		ui.PrintUsage(cmd)
		return fmt.Errorf("--format only supports %s", strings.Join(validRunFormats, ", "))
	}

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

	usingCache := make([]bool, 0, len(projectContexts))
	cacheErrors := make([]string, 0, len(projectContexts))
	for _, pCtx := range projectContexts {
		usingCache = append(usingCache, pCtx.UsingCache)
		cacheErrors = append(cacheErrors, pCtx.CacheErr)
	}
	env["usingCache"] = usingCache
	env["cacheErrors"] = cacheErrors

	summary := r.FullSummary
	env["supportedResourceCounts"] = summary.SupportedResourceCounts
	env["unsupportedResourceCounts"] = summary.UnsupportedResourceCounts
	env["noPriceResourceCounts"] = summary.NoPriceResourceCounts
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
