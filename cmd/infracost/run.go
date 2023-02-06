package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/vcs"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/prices"
	"github.com/infracost/infracost/internal/providers"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
	"github.com/infracost/infracost/internal/usage"
)

type projectJob struct {
	index      int
	projectCfg *config.Project
}

type projectResult struct {
	index      int
	ctx        *config.ProjectContext
	projectOut *projectOutput
}

func addRunFlags(cmd *cobra.Command) {
	cmd.Flags().StringSlice("terraform-var-file", nil, "Load variable files, similar to Terraform's -var-file flag. Provided files must be relative to the --path flag")
	cmd.Flags().StringSlice("terraform-var", nil, "Set value for an input variable, similar to Terraform's -var flag")
	cmd.Flags().StringP("path", "p", "", "Path to the Terraform directory or JSON/plan file")

	cmd.Flags().String("config-file", "", "Path to Infracost config file. Cannot be used with path, terraform* or usage-file flags")
	cmd.Flags().String("usage-file", "", "Path to Infracost usage file that specifies values for usage-based resources")

	cmd.Flags().String("project-name", "", "Name of project in the output. Defaults to path or git repo name")

	cmd.Flags().Bool("terraform-force-cli", false, "Generate the Terraform plan JSON using the Terraform CLI. This may require cloud credentials")
	cmd.Flags().String("terraform-plan-flags", "", "Flags to pass to 'terraform plan'. Applicable with --terraform-force-cli")
	cmd.Flags().String("terraform-init-flags", "", "Flags to pass to 'terraform init'. Applicable with --terraform-force-cli")
	cmd.Flags().String("terraform-workspace", "", "Terraform workspace to use. Applicable when path is a Terraform directory")

	cmd.Flags().StringSlice("exclude-path", nil, "Paths of directories to exclude, glob patterns need quotes")
	cmd.Flags().Bool("include-all-paths", false, "Set project auto-detection to use all subdirectories in given path")
	cmd.Flags().String("git-diff-target", "master", "Show only costs that have git changes compared to the provided branch. Use the name of the current branch to fetch changes from the last two commits")
	_ = cmd.Flags().MarkHidden("git-diff-target")

	cmd.Flags().Bool("no-cache", false, "Don't attempt to cache Terraform plans")

	cmd.Flags().Bool("show-skipped", false, "List unsupported and free resources")

	cmd.Flags().Bool("sync-usage-file", false, "Sync usage-file with missing resources, needs usage-file too (experimental)")

	_ = cmd.MarkFlagFilename("path", "json", "tf")
	_ = cmd.MarkFlagFilename("config-file", "yml")
	_ = cmd.MarkFlagFilename("usage-file", "yml")

	_ = cmd.Flags().MarkHidden("terraform-force-cli")
	// These are deprecated and will show a warning if used without --terraform-force-cli
	_ = cmd.Flags().MarkHidden("terraform-plan-flags")
	_ = cmd.Flags().MarkHidden("terraform-init-flags")
}

func runMain(cmd *cobra.Command, runCtx *config.RunContext) error {
	if runCtx.Config.IsSelfHosted() && runCtx.IsCloudEnabled() {
		ui.PrintWarning(cmd.ErrOrStderr(), "Infracost Cloud is part of Infracost's hosted services. Contact hello@infracost.io for help.")
	}

	repoPath := runCtx.Config.RepoPath()
	metadata, err := vcs.MetadataFetcher.Get(repoPath, runCtx.Config.GitDiffTarget)
	if err != nil {
		logging.Logger.WithError(err).Debugf("failed to fetch vcs metadata for path %s", repoPath)
	}
	runCtx.VCSMetadata = metadata

	pr, err := newParallelRunner(cmd, runCtx)
	if err != nil {
		return err
	}

	projectResults, err := pr.run()
	if err != nil {
		return err
	}

	projects := make([]*schema.Project, 0)
	projectContexts := make([]*config.ProjectContext, 0)

	for _, projectResult := range projectResults {
		for _, project := range projectResult.projectOut.projects {
			projectContexts = append(projectContexts, projectResult.ctx)
			projects = append(projects, project)
		}
	}

	err = checkIfAllProjectsErrored(projects)
	if err != nil {
		return err
	}

	r, err := output.ToOutputFormat(projects)
	if err != nil {
		return err
	}

	if pr.prior != nil {
		r, err = output.CompareTo(r, *pr.prior)
		if err != nil {
			return err
		}
	}

	r.IsCIRun = runCtx.IsCIRun()
	r.Currency = runCtx.Config.Currency
	r.Metadata = output.NewMetadata(runCtx)

	if runCtx.IsCloudUploadEnabled() {
		dashboardClient := apiclient.NewDashboardAPIClient(runCtx)
		result, err := dashboardClient.AddRun(runCtx, r)
		if err != nil {
			log.WithError(err).Error("Failed to upload to Infracost Cloud")
		}

		r.RunID, r.ShareURL = result.RunID, result.ShareURL
	} else {
		log.Debug("Skipping sending project results since Infracost Cloud upload is not enabled.")
	}

	format := strings.ToLower(runCtx.Config.Format)
	isCompareRun := runCtx.Config.CompareTo != ""
	if isCompareRun && !validCompareToFormats[format] {
		return errors.New("The --compare-to option cannot be used with table and html formats as they output breakdowns, specify a different --format.")
	}

	b, err := output.FormatOutput(format, r, output.Options{
		DashboardEndpoint: runCtx.Config.DashboardEndpoint,
		ShowSkipped:       runCtx.Config.ShowSkipped,
		NoColor:           runCtx.Config.NoColor,
		Fields:            runCtx.Config.Fields,
		CurrencyFormat:    runCtx.Config.CurrencyFormat,
	})
	if err != nil {
		return err
	}

	if format == "diff" || format == "table" {
		lines := bytes.Count(b, []byte("\n")) + 1
		runCtx.SetContextValue("lineCount", lines)
	}

	env := buildRunEnv(runCtx, projectContexts, r)

	pricingClient := apiclient.NewPricingAPIClient(runCtx)
	err = pricingClient.AddEvent("infracost-run", env)
	if err != nil {
		log.Errorf("Error reporting event: %s", err)
	}

	if outFile, _ := cmd.Flags().GetString("out-file"); outFile != "" {
		err = saveOutFile(runCtx, cmd, outFile, b)
		if err != nil {
			return err
		}
	} else {
		// Print a new line to separate the logs from the output
		if runCtx.Config.IsLogging() {
			cmd.PrintErrln()
		}
		cmd.Println(string(b))
	}

	return nil
}

func checkIfAllProjectsErrored(projects []*schema.Project) error {
	allError := true

	if len(projects) == 0 {
		allError = false
	}

	if len(projects) == 1 && len(projects[0].Metadata.Errors) == 1 {
		return errors.New(projects[0].Metadata.Errors[0].Message)
	}

	for _, project := range projects {
		if !project.Metadata.HasErrors() {
			allError = false
			break
		}
	}

	if allError {
		multi := &multierror.Error{}
		for _, project := range projects {
			msg := fmt.Sprintf("project %s errored: \n", project.Name)
			for i, diag := range project.Metadata.Errors {
				msg += "\t\t\t" + diag.Message

				if i != len(project.Metadata.Errors)-1 {
					msg += "\n"
				}

			}

			multi.Errors = append(multi.Errors, errors.New(msg))
		}

		return multi
	}

	return nil
}

type projectOutput struct {
	projects []*schema.Project
}

type parallelRunner struct {
	cmd         *cobra.Command
	runCtx      *config.RunContext
	pathMuxs    map[string]*sync.Mutex
	prior       *output.Root
	parallelism int
	numJobs     int
}

func newParallelRunner(cmd *cobra.Command, runCtx *config.RunContext) (*parallelRunner, error) {
	// Create a mutex for each path, so we can synchronize the runs of any
	// projects that have the same path. This is necessary because Terraform
	// can't run multiple operations in parallel on the same path.
	pathMuxs := map[string]*sync.Mutex{}
	for _, projectCfg := range runCtx.Config.Projects {
		pathMuxs[projectCfg.Path] = &sync.Mutex{}
	}

	var prior *output.Root
	if runCtx.Config.CompareTo != "" {
		snapshot, err := output.Load(runCtx.Config.CompareTo)
		if err != nil {
			return nil, fmt.Errorf("Error loading %s used by --compare-to flag. %s", runCtx.Config.CompareTo, err)
		}

		prior = &snapshot
	}

	parallelism, err := runCtx.GetParallelism()
	if err != nil {
		return nil, err
	}
	runCtx.SetContextValue("parallelism", parallelism)

	numJobs := len(runCtx.Config.Projects)

	runInParallel := parallelism > 1 && numJobs > 1
	if (runInParallel || runCtx.IsCIRun()) && !runCtx.Config.IsLogging() {
		if runInParallel {
			cmd.PrintErrln("Running multiple projects in parallel, so log-level=info is enabled by default.")
			cmd.PrintErrln("Run with INFRACOST_PARALLELISM=1 to disable parallelism to help debugging.")
			cmd.PrintErrln()
		}

		runCtx.Config.LogLevel = "info"
		err := logging.ConfigureBaseLogger(runCtx.Config)
		if err != nil {
			return nil, err
		}
	}

	return &parallelRunner{
		parallelism: parallelism,
		numJobs:     numJobs,
		runCtx:      runCtx,
		cmd:         cmd,
		pathMuxs:    pathMuxs,
		prior:       prior,
	}, nil
}

func (r *parallelRunner) run() ([]projectResult, error) {
	projectResultChan := make(chan projectResult, r.numJobs)
	jobs := make(chan projectJob, r.numJobs)

	errGroup, _ := errgroup.WithContext(context.Background())
	for i := 0; i < r.parallelism; i++ {
		i := i
		errGroup.Go(func() (err error) {
			// defer a function to recover from any panics spawned by child goroutines.
			// This is done as recover works only in the same goroutine that it is called.
			// We need to catch any child goroutine panics and hand them up to the main caller
			// so that it can be caught and displayed correctly to the user.
			defer func() {
				e := recover()
				if e != nil {
					err = clierror.NewPanicError(fmt.Errorf("%s", e), debug.Stack())
				}
			}()

			for job := range jobs {
				ctx := config.NewProjectContext(r.runCtx, job.projectCfg, log.Fields{
					"routine": i,
				})
				configProjects, err := r.runProjectConfig(ctx)
				if err != nil {
					configProjects = newErroredProject(ctx, err)
				}

				projectResultChan <- projectResult{
					index:      job.index,
					ctx:        ctx,
					projectOut: configProjects,
				}
			}

			return nil
		})
	}

	for i, p := range r.runCtx.Config.Projects {
		jobs <- projectJob{index: i, projectCfg: p}
	}
	close(jobs)

	err := errGroup.Wait()
	if err != nil {
		return nil, err
	}

	close(projectResultChan)

	projectResults := make([]projectResult, 0, len(r.runCtx.Config.Projects))
	for result := range projectResultChan {
		projectResults = append(projectResults, result)
	}

	sort.Slice(projectResults, func(i, j int) bool {
		return projectResults[i].index < projectResults[j].index
	})

	return projectResults, nil
}

func (r *parallelRunner) runProjectConfig(ctx *config.ProjectContext) (*projectOutput, error) {
	mux := r.pathMuxs[ctx.ProjectConfig.Path]
	if mux != nil {
		mux.Lock()
		defer mux.Unlock()
	}

	provider, err := providers.Detect(ctx, r.prior == nil)
	var warn *string
	if v, ok := err.(*providers.ValidationError); ok {
		if v.Warn() == nil {
			return nil, err
		}

		warn = v.Warn()
	} else if err != nil {
		m := fmt.Sprintf("%s\n\n", err)
		m += fmt.Sprintf("Try setting --path to a Terraform plan JSON file. See %s for how to generate this.", ui.LinkString("https://infracost.io/troubleshoot"))

		return nil, clierror.NewCLIError(errors.New(m), "Could not detect path type")
	}

	ctx.SetContextValue("projectType", provider.Type())

	projectTypes := []interface{}{}
	if t, ok := ctx.RunContext.ContextValues()["projectTypes"]; ok {
		projectTypes = t.([]interface{})
	}
	projectTypes = append(projectTypes, provider.Type())
	ctx.RunContext.SetContextValue("projectTypes", projectTypes)

	if r.cmd.Name() == "diff" && provider.Type() == "terraform_state_json" {
		m := "Cannot use Terraform state JSON with the infracost diff command.\n\n"
		m += fmt.Sprintf("Use the %s flag to specify the path to one of the following:\n", ui.PrimaryString("--path"))
		m += fmt.Sprintf(" - Terraform/Terragrunt directory\n - Terraform plan JSON file, see %s for how to generate this.", ui.SecondaryLinkString("https://infracost.io/troubleshoot"))
		return nil, clierror.NewCLIError(errors.New(m), "Cannot use Terraform state JSON with the infracost diff command")
	}

	m := fmt.Sprintf("Detected %s at %s", provider.DisplayType(), ui.DisplayPath(ctx.ProjectConfig.Path))
	if provider.Type() == "terraform_dir" {
		m = fmt.Sprintf("Evaluating %s at %s", provider.DisplayType(), ui.DisplayPath(ctx.ProjectConfig.Path))
	}

	if r.runCtx.Config.IsLogging() {
		log.Info(m)
	} else {
		fmt.Fprintln(os.Stderr, m)
	}

	if warn != nil {
		ui.PrintWarning(r.runCtx.ErrWriter, *warn)
	}

	// Generate usage file
	if r.runCtx.Config.SyncUsageFile {
		err := r.generateUsageFile(ctx, provider)
		if err != nil {
			return nil, errors.Wrap(err, "Error generating usage file")
		}
	}

	// Load usage data
	usageData := make(map[string]*schema.UsageData)
	var usageFile *usage.UsageFile

	if ctx.ProjectConfig.UsageFile != "" {
		var err error
		usageFile, err = usage.LoadUsageFile(ctx.ProjectConfig.UsageFile)
		if err != nil {
			return nil, err
		}

		invalidKeys, err := usageFile.InvalidKeys()
		if err != nil {
			log.Errorf("Error checking usage file keys: %v", err)
		} else if len(invalidKeys) > 0 {
			ui.PrintWarningf(r.cmd.ErrOrStderr(),
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
	out := &projectOutput{}

	t1 := time.Now()
	projects, err := provider.LoadResources(usageData)
	if err != nil {
		r.cmd.PrintErrln()
		return nil, err
	}

	_ = r.uploadCloudResourceIDs(projects)

	r.buildResources(projects)

	spinnerOpts := ui.SpinnerOptions{
		EnableLogging: r.runCtx.Config.IsLogging(),
		NoColor:       r.runCtx.Config.NoColor,
		Indent:        "  ",
	}
	spinner := ui.NewSpinner("Retrieving cloud prices to calculate costs", spinnerOpts)
	defer spinner.Fail()

	for _, project := range projects {
		if err := prices.PopulatePrices(r.runCtx, project); err != nil {
			spinner.Fail()
			r.cmd.PrintErrln()

			if e := unwrapped(err); errors.Is(e, apiclient.ErrInvalidAPIKey) {
				return nil, fmt.Errorf("%v\n%s %s %s %s %s\n%s %s.\n%s %s %s",
					e.Error(),
					"Please check your",
					ui.PrimaryString(config.CredentialsFilePath()),
					"file or",
					ui.PrimaryString("INFRACOST_API_KEY"),
					"environment variable.",
					"If you recently regenerated your API key, you can retrieve it from",
					ui.PrimaryString(r.runCtx.Config.DashboardEndpoint),
					"See",
					ui.PrimaryString("https://infracost.io/support"),
					"if you continue having issues.",
				)
			}

			if e, ok := err.(*apiclient.APIError); ok {
				return nil, fmt.Errorf("%v\n%s", e.Error(), "We have been notified of this issue.")
			}

			return nil, err
		}
		schema.CalculateCosts(project)

		project.CalculateDiff()
	}

	t2 := time.Now()
	taken := t2.Sub(t1).Milliseconds()
	ctx.SetContextValue("tfProjectRunTimeMs", taken)

	spinner.Success()

	if r.runCtx.Config.UsageActualCosts {
		r.populateActualCosts(projects)
	}

	out.projects = projects

	if !r.runCtx.Config.IsLogging() && !r.runCtx.Config.SkipErrLine {
		r.cmd.PrintErrln()
	}

	return out, nil
}

func (r *parallelRunner) uploadCloudResourceIDs(projects []*schema.Project) error {
	if r.runCtx.Config.UsageAPIEndpoint == "" || !r.hasCloudResourceIDToUpload(projects) {
		return nil
	}

	r.runCtx.SetContextValue("uploadedResourceIds", true)

	spinnerOpts := ui.SpinnerOptions{
		EnableLogging: r.runCtx.Config.IsLogging(),
		NoColor:       r.runCtx.Config.NoColor,
		Indent:        "  ",
	}
	spinner := ui.NewSpinner("Sending resource IDs to Infracost Cloud for usage estimates", spinnerOpts)
	defer spinner.Fail()

	for _, project := range projects {
		if err := prices.UploadCloudResourceIDs(r.runCtx, project); err != nil {
			logging.Logger.WithError(err).Debugf("failed to upload resource IDs for project %s", project.Name)
			return err
		}
	}

	spinner.Success()
	return nil
}

func (r *parallelRunner) hasCloudResourceIDToUpload(projects []*schema.Project) bool {
	for _, project := range projects {
		for _, partial := range project.AllPartialResources() {
			if len(partial.CloudResourceIDs) > 0 {
				return true
			}
		}
	}

	return false
}

func (r *parallelRunner) buildResources(projects []*schema.Project) {
	var projectPtrToUsageMap map[*schema.Project]map[string]*schema.UsageData
	if r.runCtx.Config.UsageAPIEndpoint != "" {
		projectPtrToUsageMap = r.fetchProjectUsage(projects)
	}

	schema.BuildResources(projects, projectPtrToUsageMap)
}

func (r *parallelRunner) fetchProjectUsage(projects []*schema.Project) map[*schema.Project]map[string]*schema.UsageData {
	coreResourceCount := 0
	for _, project := range projects {
		for _, partial := range project.PartialResources {
			if partial.CoreResource != nil {
				coreResourceCount++
			}
		}
	}

	if coreResourceCount == 0 {
		return nil
	}

	resourceStr := fmt.Sprintf("%d resource", coreResourceCount)
	if coreResourceCount > 1 {
		resourceStr += "s"
	}

	spinnerOpts := ui.SpinnerOptions{
		EnableLogging: r.runCtx.Config.IsLogging(),
		NoColor:       r.runCtx.Config.NoColor,
		Indent:        "  ",
	}
	spinner := ui.NewSpinner(fmt.Sprintf("Retrieving usage estimates for %s from Infracost Cloud", resourceStr), spinnerOpts)
	defer spinner.Fail()

	projectPtrToUsageMap := make(map[*schema.Project]map[string]*schema.UsageData, len(projects))

	for _, project := range projects {
		usageMap, err := prices.FetchUsageData(r.runCtx, project)
		if err != nil {
			logging.Logger.WithError(err).Debugf("failed to retrieve usage data for project %s", project.Name)
			return nil
		}
		r.runCtx.SetContextValue("fetchedUsageData", true)
		projectPtrToUsageMap[project] = usageMap
	}

	spinner.Success()

	return projectPtrToUsageMap
}

func (r *parallelRunner) populateActualCosts(projects []*schema.Project) {
	if r.runCtx.Config.UsageAPIEndpoint != "" {
		spinnerOpts := ui.SpinnerOptions{
			EnableLogging: r.runCtx.Config.IsLogging(),
			NoColor:       r.runCtx.Config.NoColor,
			Indent:        "  ",
		}
		spinner := ui.NewSpinner("Retrieving actual costs from Infracost Cloud", spinnerOpts)
		defer spinner.Fail()

		for _, project := range projects {
			if err := prices.PopulateActualCosts(r.runCtx, project); err != nil {
				logging.Logger.WithError(err).Debugf("failed to retrieve actual costs for project %s", project.Name)
				return
			}
		}

		spinner.Success()
	}
}

func (r *parallelRunner) generateUsageFile(ctx *config.ProjectContext, provider schema.Provider) error {
	if ctx.ProjectConfig.UsageFile == "" {
		// This should not happen as we check earlier in the code that usage-file is not empty when sync-usage-file flag is on.
		return fmt.Errorf("Error generating usage: no usage file given")
	}

	var usageFile *usage.UsageFile

	usageFilePath := ctx.ProjectConfig.UsageFile
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

	r.buildResources(providerProjects)

	spinnerOpts := ui.SpinnerOptions{
		EnableLogging: r.runCtx.Config.IsLogging(),
		NoColor:       r.runCtx.Config.NoColor,
		Indent:        "  ",
	}

	spinner := ui.NewSpinner("Syncing usage data from cloud", spinnerOpts)
	defer spinner.Fail()

	syncResult, err := usage.SyncUsageData(ctx, usageFile, providerProjects)

	if err != nil {
		spinner.Fail()
		return errors.Wrap(err, "Error synchronizing usage data")
	}

	ctx.SetFrom(syncResult)
	if err != nil {
		spinner.Fail()
		return errors.Wrap(err, "Error summarizing usage")
	}

	err = usageFile.WriteToPath(ctx.ProjectConfig.UsageFile)
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
		r.cmd.PrintErrln(fmt.Sprintf("    %s Synced %d of %d resource%s",
			ui.FaintString("└─"),
			successes,
			resources,
			pluralized))
	}
	return nil
}

func loadRunFlags(cfg *config.Config, cmd *cobra.Command) error {
	hasPathFlag := cmd.Flags().Changed("path")
	hasConfigFile := cmd.Flags().Changed("config-file")

	if cmd.Flags().Changed("git-diff-target") {
		s, _ := cmd.Flags().GetString("git-diff-target")
		cfg.GitDiffTarget = &s
	}

	cfg.CompareTo, _ = cmd.Flags().GetString("compare-to")

	cfg.CompareTo, _ = cmd.Flags().GetString("compare-to")

	if cmd.Name() != "infracost" && !hasPathFlag && !hasConfigFile {
		m := fmt.Sprintf("No path specified\n\nUse the %s flag to specify the path to one of the following:\n", ui.PrimaryString("--path"))
		m += fmt.Sprintf(" - Terraform/Terragrunt directory\n - Terraform plan JSON file, see %s for how to generate this.", ui.SecondaryLinkString("https://infracost.io/troubleshoot"))
		m += fmt.Sprintf("\n\nAlternatively, use --config-file to process multiple projects, see %s", ui.SecondaryLinkString("https://infracost.io/config-file"))

		ui.PrintUsage(cmd)
		return errors.New(m)
	}

	hasProjectFlags := (hasPathFlag ||
		cmd.Flags().Changed("usage-file") ||
		cmd.Flags().Changed("project-name") ||
		cmd.Flags().Changed("terraform-plan-flags") ||
		cmd.Flags().Changed("terraform-var-file") ||
		cmd.Flags().Changed("terraform-var") ||
		cmd.Flags().Changed("terraform-init-flags") ||
		cmd.Flags().Changed("terraform-workspace"))

	if hasConfigFile && hasProjectFlags {
		m := "--config-file flag cannot be used with the following flags: "
		m += "--path, --project-name, --terraform-*, --usage-file"
		ui.PrintUsage(cmd)
		return errors.New(m)
	}

	projectCfg := cfg.Projects[0]

	if hasProjectFlags {
		rootPath, _ := cmd.Flags().GetString("path")
		cfg.RootPath = rootPath
		projectCfg.Path = rootPath

		projectCfg.TerraformVarFiles, _ = cmd.Flags().GetStringSlice("terraform-var-file")
		tfVars, _ := cmd.Flags().GetStringSlice("terraform-var")
		projectCfg.TerraformVars = tfVarsToMap(tfVars)
		projectCfg.UsageFile, _ = cmd.Flags().GetString("usage-file")
		projectCfg.Name, _ = cmd.Flags().GetString("project-name")
		projectCfg.TerraformForceCLI, _ = cmd.Flags().GetBool("terraform-force-cli")
		projectCfg.TerraformPlanFlags, _ = cmd.Flags().GetString("terraform-plan-flags")
		projectCfg.TerraformInitFlags, _ = cmd.Flags().GetString("terraform-init-flags")
		projectCfg.TerraformUseState, _ = cmd.Flags().GetBool("terraform-use-state")
		projectCfg.ExcludePaths, _ = cmd.Flags().GetStringSlice("exclude-path")
		projectCfg.IncludeAllPaths, _ = cmd.Flags().GetBool("include-all-paths")

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

		cfg.ConfigFilePath = cfgFilePath

		if forceCLI, _ := cmd.Flags().GetBool("terraform-force-cli"); forceCLI {
			for _, p := range cfg.Projects {
				p.TerraformForceCLI = true
			}
		}
		if useState, _ := cmd.Flags().GetBool("terraform-use-state"); useState {
			for _, p := range cfg.Projects {
				p.TerraformUseState = true
			}
		}
	}

	cfg.NoCache, _ = cmd.Flags().GetBool("no-cache")
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

func tfVarsToMap(vars []string) map[string]string {
	if len(vars) == 0 {
		return nil
	}

	m := make(map[string]string, len(vars))
	for _, v := range vars {
		pieces := strings.Split(v, "=")
		if len(pieces) != 2 {
			continue
		}

		m[pieces[0]] = pieces[1]
	}

	return m
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

	env["runId"] = r.RunID
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

	if warnings := runCtx.GetResourceWarnings(); warnings != nil {
		env["resourceWarnings"] = warnings
	}

	if n := r.ExampleProjectName(); n != "" {
		env["exampleProjectName"] = n
	}

	return env
}

func unwrapped(err error) error {
	e := err
	for errors.Unwrap(e) != nil {
		e = errors.Unwrap(e)
	}

	return e
}

func newErroredProject(ctx *config.ProjectContext, err error) *projectOutput {
	metadata := config.DetectProjectMetadata(ctx.ProjectConfig.Path)
	metadata.Type = "error"
	metadata.AddError(err)

	name := ctx.ProjectConfig.Name
	if name == "" {
		name = metadata.GenerateProjectName(ctx.RunContext.VCSMetadata.Remote, ctx.RunContext.IsCloudEnabled())
	}

	return &projectOutput{projects: []*schema.Project{schema.NewProject(name, metadata)}}
}
