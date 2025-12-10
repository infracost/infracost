package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/infracost/infracost/internal/metrics"
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
	index    int
	provider schema.Provider
	err      error
	ctx      *config.ProjectContext
}

type projectResult struct {
	index      int
	ctx        *config.ProjectContext
	projectOut *projectOutput
}

func addRunFlags(cmd *cobra.Command) {
	cmd.Flags().StringSlice("terraform-var-file", nil, "Load variable files, similar to Terraform's -var-file flag. Provided files must be relative to the --path flag")
	cmd.Flags().StringArray("terraform-var", nil, "Set value for an input variable, similar to Terraform's -var flag")
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

	cmd.Flags().Bool("show-skipped", false, "List unsupported resources")

	cmd.Flags().Bool("sync-usage-file", false, "Sync usage-file with missing resources, needs usage-file too (experimental)")

	_ = cmd.MarkFlagFilename("path", "json", "tf", "tofu")
	_ = cmd.MarkFlagFilename("config-file", "yml")
	_ = cmd.MarkFlagFilename("usage-file", "yml")

	_ = cmd.Flags().MarkHidden("terraform-force-cli")
	// These are deprecated and will show a warning if used without --terraform-force-cli
	_ = cmd.Flags().MarkHidden("terraform-plan-flags")
	_ = cmd.Flags().MarkHidden("terraform-init-flags")
}

func runMain(cmd *cobra.Command, runCtx *config.RunContext) error {
	if runCtx.Config.IsSelfHosted() && runCtx.IsCloudEnabled() {
		logging.Logger.Warn().Msg("Infracost Cloud is part of Infracost's hosted services. Contact hello@infracost.io for help.")
	}

	wd := runCtx.Config.WorkingDirectory()
	metadata, err := vcs.MetadataFetcher.Get(wd, runCtx.Config.GitDiffTarget)
	if err != nil {
		logging.Logger.Debug().Err(err).Msgf("failed to fetch vcs metadata for path %s", wd)
	}
	runCtx.VCSMetadata = metadata
	runCtx.VCSMetadata.BaseCommit = vcs.Commit{}

	pr, err := newParallelRunner(cmd, runCtx)
	if err != nil {
		return err
	}

	projectResults, err := pr.run()
	if err != nil {
		return err
	}

	// write an aggregate log line of cost components that have
	// missing prices if any have been found.
	pr.pricingFetcher.LogWarnings()

	projects := make([]*schema.Project, 0)
	projectContexts := make([]*config.ProjectContext, 0)

	for _, projectResult := range projectResults {
		projectContexts = append(projectContexts, projectResult.ctx)
		projects = append(projects, projectResult.projectOut.projects...)
	}

	r, err := output.ToOutputFormat(runCtx.Config, projects)
	if err != nil {
		return err
	}

	if pr.prior != nil {
		r, err = output.CompareTo(runCtx.Config, r, *pr.prior)
		if err != nil {
			return err
		}
		runCtx.VCSMetadata.BaseCommit = vcs.Commit{
			SHA:         pr.prior.Metadata.CommitSHA,
			AuthorName:  pr.prior.Metadata.CommitAuthorName,
			AuthorEmail: pr.prior.Metadata.CommitAuthorEmail,
			Time:        pr.prior.Metadata.CommitTimestamp,
			Message:     pr.prior.Metadata.CommitMessage,
		}
	}

	r.IsCIRun = runCtx.IsCIRun()
	r.Currency = runCtx.Config.Currency
	r.Metadata = output.NewMetadata(runCtx)

	if runCtx.IsCloudUploadExplicitlyEnabled() {
		dashboardClient := apiclient.NewDashboardAPIClient(runCtx)
		result, err := dashboardClient.AddRun(runCtx, r)
		if err != nil {
			logging.Logger.Err(err).Msg("Failed to upload to Infracost Cloud")
		}

		r.RunID, r.ShareURL, r.CloudURL = result.RunID, result.ShareURL, result.CloudURL
	} else {
		logging.Logger.Debug().Msg("Skipping sending project results since Infracost Cloud upload is not enabled.")
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
		runCtx.ContextValues.SetValue("lineCount", lines)
	}

	env := pr.buildRunEnv(projectContexts, r)

	pricingClient := apiclient.GetPricingAPIClient(runCtx)
	err = pricingClient.AddEvent("infracost-run", env)
	if err != nil {
		logging.Logger.Error().Msgf("Error reporting event: %s", err)
	}

	if outFile, _ := cmd.Flags().GetString("out-file"); outFile != "" {
		err = saveOutFile(runCtx, cmd, outFile, b)
		if err != nil {
			return err
		}
	} else {
		// Print a new line to separate the logs from the output
		cmd.PrintErrln()
		cmd.Println(string(b))
	}

	return nil
}

type projectOutput struct {
	projects []*schema.Project
}

type parallelRunner struct {
	cmd            *cobra.Command
	runCtx         *config.RunContext
	pathMuxs       map[string]*sync.Mutex
	prior          *output.Root
	parallelism    int
	pricingFetcher *prices.PriceFetcher
}

func newParallelRunner(cmd *cobra.Command, runCtx *config.RunContext) (*parallelRunner, error) {
	// Create a mutex for each path, so we can synchronize the runs of any
	// projects that have the same path. This is necessary because Terraform
	// can't run multiple operations in parallel on the same path.
	pathMuxs := map[string]*sync.Mutex{}
	for _, projectCfg := range runCtx.Config.Projects {
		if projectCfg.TerraformForceCLI {
			pathMuxs[projectCfg.Path] = &sync.Mutex{}
		}
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
	runCtx.ContextValues.SetValue("parallelism", parallelism)

	metrics.GetCounter("parallel_runner.parallelism", false).Add(parallelism)

	return &parallelRunner{
		parallelism:    parallelism,
		runCtx:         runCtx,
		cmd:            cmd,
		pathMuxs:       pathMuxs,
		prior:          prior,
		pricingFetcher: prices.NewPriceFetcher(runCtx, false),
	}, nil
}

func (r *parallelRunner) run() ([]projectResult, error) {
	var queue []projectJob
	var totalRootModules int
	var i int

	parallelRunnerTimer := metrics.GetTimer("parallel_runner.run.total_duration", false).Start()
	defer parallelRunnerTimer.Stop()

	isAuto := r.runCtx.IsAutoDetect()
	for _, p := range r.runCtx.Config.Projects {
		detectionOutput, err := providers.Detect(r.runCtx, p, r.prior == nil)
		if err != nil {
			m := fmt.Sprintf("%s\n\n", err)
			m += fmt.Sprintf("  Try adding a config-file to configure how Infracost should run. See %s for details and examples.", ui.LinkString("https://infracost.io/config-file"))

			queue = append(queue, projectJob{index: i, err: schema.NewEmptyPathTypeError(errors.New(m)), ctx: config.NewProjectContext(r.runCtx, p, map[string]interface{}{})})
			i++
			continue
		}

		for _, provider := range detectionOutput.Providers {
			queue = append(queue, projectJob{index: i, provider: provider})
			i++
		}

		totalRootModules += detectionOutput.RootModules
	}

	metrics.GetCounter("parallel_runner.project_count", false).Add(i)
	metrics.GetCounter("parallel_runner.root_module_count", false).Add(totalRootModules)

	projectCounts := make(map[string]int)
	for _, job := range queue {
		if job.err != nil {
			continue
		}

		provider := job.provider
		if v, ok := projectCounts[provider.DisplayType()]; ok {
			projectCounts[provider.DisplayType()] = v + 1
			continue
		}

		projectCounts[provider.DisplayType()] = 1
	}

	order := make([]string, 0, len(projectCounts))
	for displayType := range projectCounts {
		order = append(order, displayType)
	}

	var summary string
	sort.Strings(order)
	for i, displayType := range order {
		count := projectCounts[displayType]
		desc := "project"
		if count > 1 {
			desc = "projects"
		}

		if len(order) > 1 && i == len(order)-2 {
			summary += fmt.Sprintf("%d %s %s and ", count, displayType, desc)
		} else if i == len(order)-1 {
			summary += fmt.Sprintf("%d %s %s", count, displayType, desc)
		} else {
			summary += fmt.Sprintf("%d %s %s, ", count, displayType, desc)
		}
	}

	moduleDesc := "module"
	pathDesc := "path"
	if totalRootModules > 1 {
		moduleDesc = "modules"
	}

	if len(r.runCtx.Config.Projects) > 1 {
		pathDesc = "paths"
	}

	if isAuto {
		if summary == "" {
			logging.Logger.Error().Msgf("Could not autodetect any projects from path %s", ui.DirectoryDisplayName(r.runCtx, r.runCtx.Config.RootPath))
		} else {
			logging.Logger.Info().Msgf("Autodetected %s across %d root %s", summary, totalRootModules, moduleDesc)
		}
	} else {
		if summary == "" {
			logging.Logger.Error().Msg("All provided config file paths are invalid or do not contain any supported projects")
		} else {
			logging.Logger.Info().Msgf("Autodetected %s from %d %s in the config file", summary, len(r.runCtx.Config.Projects), pathDesc)
		}
	}

	for _, job := range queue {
		if job.err != nil {
			continue
		}

		provider := job.provider

		name := provider.ProjectName()
		displayName := ui.ProjectDisplayName(r.runCtx, name)

		dirDisp := ui.DirectoryDisplayName(r.runCtx, provider.RelativePath())
		if len(provider.VarFiles()) > 0 {
			varString := ""
			for _, s := range provider.VarFiles() {
				varString += fmt.Sprintf("%s, ", ui.DirectoryDisplayName(r.runCtx, filepath.Join(provider.RelativePath(), s)))
			}
			varString = strings.TrimRight(varString, ", ")

			logging.Logger.Info().Msgf("Found %s project %s at directory %s using %s var files %v", provider.DisplayType(), displayName, dirDisp, provider.DisplayType(), varString)
		} else {
			logging.Logger.Info().Msgf("Found %s project %s at directory %s", provider.DisplayType(), displayName, dirDisp)
		}
	}

	projectResultChan := make(chan projectResult, len(queue))
	jobs := make(chan projectJob, len(queue))

	errGroup, _ := errgroup.WithContext(context.Background())
	for i := 0; i < r.parallelism; i++ {
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
				func() {
					var metricContext []string
					if job.provider != nil {
						metricContext = append(metricContext, job.provider.Type())
						if job.provider.Context() != nil {
							metricContext = append(metricContext, job.provider.Context().ProjectConfig.Path)
						}
					}
					jobTimer := metrics.GetTimer("parallel_runner.job.duration", false, metricContext...).Start()
					defer jobTimer.Stop()

					var configProjects *projectOutput
					ctx := job.ctx
					if job.err != nil {
						configProjects = newErroredProject(job.provider, job.ctx, job.err)
					} else {
						configProjects, err = r.runProvider(job)
						ctx = job.provider.Context()
						if err != nil {
							configProjects = newErroredProject(job.provider, ctx, err)
						}

					}

					projectResultChan <- projectResult{
						index:      job.index,
						ctx:        ctx,
						projectOut: configProjects,
					}
				}()
			}

			return nil
		})
	}

	allJobsTimer := metrics.GetTimer("parallel_runner.all_jobs.duration", false).Start()

	for _, job := range queue {
		jobs <- job
	}

	close(jobs)

	err := errGroup.Wait()
	allJobsTimer.Stop()
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

func (r *parallelRunner) runProvider(job projectJob) (out *projectOutput, err error) {
	projectContext := job.provider.Context()
	path := projectContext.ProjectConfig.Path
	mux := r.pathMuxs[path]
	if mux != nil {
		mux.Lock()
		defer mux.Unlock()
	}

	if r.cmd.Name() == "diff" && job.provider.Type() == "terraform_state_json" {
		m := "Cannot use Terraform state JSON with the infracost diff command.\n\n"
		m += fmt.Sprintf("Use the %s flag to specify the path to one of the following:\n", ui.PrimaryString("--path"))
		m += fmt.Sprintf(" - Terraform/Terragrunt directory\n - Terraform plan JSON file, see %s for how to generate this.", ui.SecondaryLinkString("https://infracost.io/troubleshoot"))
		return nil, clierror.NewCLIError(errors.New(m), "Cannot use Terraform state JSON with the infracost diff command")
	}

	name := job.provider.ProjectName()
	displayName := ui.ProjectDisplayName(r.runCtx, name)

	logging.Logger.Debug().Msgf("Starting evaluation for project %s", displayName)
	defer func() {
		if err != nil {
			logging.Logger.Debug().Msgf("Failed evaluation for project %s", displayName)
		}
		logging.Logger.Debug().Msgf("Finished evaluation for project %s", displayName)
	}()

	// Generate usage file
	if r.runCtx.Config.SyncUsageFile {
		usageGenTimer := metrics.GetTimer("parallel_runner.usage_gen.duration", false, path).Start()
		err = r.generateUsageFile(job.provider)
		usageGenTimer.Stop()
		if err != nil {
			return nil, fmt.Errorf("Error generating usage file %w", err)
		}
	}

	// Load usage data
	var usageFile *usage.UsageFile

	if projectContext.ProjectConfig.UsageFile != "" {
		var err error
		usageFile, err = usage.LoadUsageFile(projectContext.ProjectConfig.UsageFile)
		if err != nil {
			return nil, err
		}

		invalidKeys, err := usageFile.InvalidKeys()
		if err != nil {
			logging.Logger.Error().Msgf("Error checking usage file keys: %v", err)
		} else if len(invalidKeys) > 0 {
			logging.Logger.Warn().Msgf(
				"The following usage file parameters are invalid and will be ignored: %s\n",
				strings.Join(invalidKeys, ", "),
			)
		}

		projectContext.ContextValues.SetValue("hasUsageFile", true)
	} else {
		usageFile = usage.NewBlankUsageFile()
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

	usageData := usageFile.ToUsageDataMap()
	out = &projectOutput{}

	t1 := time.Now()
	loadResourcesTimer := metrics.GetTimer("parallel_runner.load_resources.duration", false, path).Start()
	projects, err := job.provider.LoadResources(usageData)
	loadResourcesTimer.Stop()
	if err != nil {
		r.cmd.PrintErrln()
		return nil, err
	}

	_ = r.uploadCloudResourceIDs(projects)

	buildResourcesTimer := metrics.GetTimer("parallel_runner.build_resources.duration", false, path).Start()
	r.buildResources(projects)
	buildResourcesTimer.Stop()

	costingTimer := metrics.GetTimer("parallel_runner.costing.duration", false, path).Start()
	defer costingTimer.Stop()
	logging.Logger.Debug().Msg("Retrieving cloud prices to calculate costs")

	for _, project := range projects {
		if err = r.pricingFetcher.PopulatePrices(project); err != nil {
			logging.Logger.Debug().Err(err).Msgf("failed to populate prices for project %s", project.Name)
			r.cmd.PrintErrln()

			var apiErr *apiclient.APIError
			if errors.As(err, &apiErr) {
				switch apiErr.ErrorCode {
				case apiclient.ErrorCodeExceededQuota:
					return nil, schema.NewDiagRunQuotaExceeded(apiErr)
				case apiclient.ErrorCodeAPIKeyInvalid:
					return nil, fmt.Errorf("%v\n%s %s %s %s %s\n%s %s.\n%s %s %s",
						apiErr.Msg,
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

				return nil, fmt.Errorf("%v\n%s", apiErr.Error(), "We have been notified of this issue.")
			}

			return nil, err
		}
		schema.CalculateCosts(project)

		project.CalculateDiff()
	}

	t2 := time.Now()
	taken := t2.Sub(t1).Milliseconds()
	projectContext.ContextValues.SetValue("tfProjectRunTimeMs", taken)

	if r.runCtx.Config.UsageActualCosts {
		r.populateActualCosts(projects)
	}

	out.projects = projects

	return out, nil
}

func (r *parallelRunner) uploadCloudResourceIDs(projects []*schema.Project) error {
	if r.runCtx.Config.UsageAPIEndpoint == "" || !r.hasCloudResourceIDToUpload(projects) {
		return nil
	}

	r.runCtx.ContextValues.SetValue("uploadedResourceIds", true)

	logging.Logger.Debug().Msg("Sending resource IDs to Infracost Cloud for usage estimates")

	for _, project := range projects {
		if err := prices.UploadCloudResourceIDs(r.runCtx, project); err != nil {
			logging.Logger.Debug().Err(err).Msgf("failed to upload resource IDs for project %s", project.Name)
			return err
		}
	}

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
	var projectPtrToUsageMap map[*schema.Project]schema.UsageMap
	if r.runCtx.Config.UsageAPIEndpoint != "" {
		projectPtrToUsageMap = r.fetchProjectUsage(projects)
	}

	schema.BuildResources(projects, projectPtrToUsageMap)
}

func (r *parallelRunner) fetchProjectUsage(projects []*schema.Project) map[*schema.Project]schema.UsageMap {
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

	logging.Logger.Debug().Msgf("Retrieving usage defaults for %s from Infracost Cloud", resourceStr)

	projectPtrToUsageMap := make(map[*schema.Project]schema.UsageMap, len(projects))

	for _, project := range projects {
		usageMap, err := prices.FetchUsageData(r.runCtx, project)
		if err != nil {
			logging.Logger.Debug().Err(err).Msgf("failed to retrieve usage data for project %s", project.Name)
			return nil
		}
		r.runCtx.ContextValues.SetValue("fetchedUsageData", true)
		projectPtrToUsageMap[project] = usageMap
	}

	return projectPtrToUsageMap
}

func (r *parallelRunner) populateActualCosts(projects []*schema.Project) {
	if r.runCtx.Config.UsageAPIEndpoint != "" {
		logging.Logger.Debug().Msg("Retrieving actual costs from Infracost Cloud")

		for _, project := range projects {
			if err := prices.PopulateActualCosts(r.runCtx, project); err != nil {
				logging.Logger.Debug().Err(err).Msgf("failed to retrieve actual costs for project %s", project.Name)
				return
			}
		}
	}
}

func (r *parallelRunner) generateUsageFile(provider schema.Provider) (err error) {
	ctx := provider.Context()

	if ctx.ProjectConfig.UsageFile == "" {
		// This should not happen as we check earlier in the code that usage-file is not empty when sync-usage-file flag is on.
		return fmt.Errorf("Error generating usage: no usage file given")
	}

	var usageFile *usage.UsageFile

	usageFilePath := ctx.ProjectConfig.UsageFile
	err = usage.CreateUsageFile(usageFilePath)
	if err != nil {
		return fmt.Errorf("Error creating usage file %w", err)
	}

	usageFile, err = usage.LoadUsageFile(usageFilePath)
	if err != nil {
		return fmt.Errorf("Error loading usage file %w", err)
	}

	usageData := usageFile.ToUsageDataMap()
	providerProjects, err := provider.LoadResources(usageData)
	if err != nil {
		return fmt.Errorf("Error loading resources %w", err)
	}

	r.buildResources(providerProjects)

	logging.Logger.Debug().Msg("Syncing usage data from cloud")
	defer func() {
		if err != nil {
			logging.Logger.Debug().Err(err).Msg("Error syncing usage data")
		} else {
			logging.Logger.Debug().Msg("Finished syncing usage data")
		}
	}()

	syncResult, err := usage.SyncUsageData(ctx, usageFile, providerProjects)

	if err != nil {
		return fmt.Errorf("Error synchronizing usage data %w", err)
	}

	ctx.SetFrom(syncResult)
	if err != nil {
		return fmt.Errorf("Error summarizing usage %w", err)
	}

	err = usageFile.WriteToPath(ctx.ProjectConfig.UsageFile)
	if err != nil {
		return fmt.Errorf("Error writing usage file %w", err)
	}

	if syncResult != nil {
		resources := syncResult.ResourceCount
		attempts := syncResult.EstimationCount
		errors := len(syncResult.EstimationErrors)
		successes := attempts - errors

		pluralized := ""
		if resources > 1 {
			pluralized = "s"
		}

		logging.Logger.Info().Msgf("Synced %d of %d resource%s", successes, resources, pluralized)
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
		cmd.Flags().Changed("project-name") ||
		cmd.Flags().Changed("terraform-plan-flags") ||
		cmd.Flags().Changed("terraform-var-file") ||
		cmd.Flags().Changed("terraform-var") ||
		cmd.Flags().Changed("terraform-init-flags") ||
		cmd.Flags().Changed("terraform-workspace"))

	if hasConfigFile && hasProjectFlags {
		m := "--config-file flag cannot be used with the following flags: "
		m += "--path, --project-name, --terraform-*"
		ui.PrintUsage(cmd)
		return errors.New(m)
	}

	projectCfg := cfg.Projects[0]

	if hasProjectFlags {
		rootPath, _ := cmd.Flags().GetString("path")
		cfg.RootPath = rootPath
		projectCfg.Path = rootPath

		projectCfg.TerraformVarFiles, _ = cmd.Flags().GetStringSlice("terraform-var-file")
		tfVars, _ := cmd.Flags().GetStringArray("terraform-var")
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
		err := cfg.LoadFromConfigFile(cfgFilePath, cmd)

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
		if usageFilePath, _ := cmd.Flags().GetString("usage-file"); usageFilePath != "" {
			for _, p := range cfg.Projects {
				if p.UsageFile == "" {
					p.UsageFile = usageFilePath
				}
			}
		}
	}

	cfg.NoCache, _ = cmd.Flags().GetBool("no-cache")
	cfg.Format, _ = cmd.Flags().GetString("format")
	cfg.ShowSkipped, _ = cmd.Flags().GetBool("show-skipped")
	cfg.SyncUsageFile, _ = cmd.Flags().GetBool("sync-usage-file")
	cfg.UsageFilePath, _ = cmd.Flags().GetString("usage-file")

	includeAllFields := "all"
	validFields := []string{"price", "monthlyQuantity", "unit", "hourlyCost", "monthlyCost"}
	validFieldsFormats := []string{"table", "html"}

	if cmd.Flags().Changed("fields") {
		fields, _ := cmd.Flags().GetStringSlice("fields")
		if len(fields) == 0 {
			logging.Logger.Warn().Msgf("fields is empty, using defaults: %s", cmd.Flag("fields").DefValue)
		} else if cfg.Fields != nil && !contains(validFieldsFormats, cfg.Format) {
			logging.Logger.Warn().Msg("fields is only supported for table and html output formats")
		} else if len(fields) == 1 && fields[0] == includeAllFields {
			cfg.Fields = validFields
		} else {
			vf := []string{}
			for _, f := range fields {
				if !contains(validFields, f) {
					logging.Logger.Warn().Msgf("Invalid field '%s' specified, valid fields are: %s or '%s' to include all fields", f, validFields, includeAllFields)
				} else {
					vf = append(vf, f)
				}
			}
			cfg.Fields = vf
		}
	}

	return nil
}

func tfVarsToMap(vars []string) map[string]interface{} {
	if len(vars) == 0 {
		return nil
	}

	m := make(map[string]interface{}, len(vars))
	for _, v := range vars {
		pieces := strings.SplitN(v, "=", 2)

		if len(pieces) != 2 {
			continue
		}

		var v interface{}
		err := json.Unmarshal([]byte(pieces[1]), &v)
		if err != nil {
			// If there's an error it could just be a raw string value, so we use that
			v = pieces[1]
		}

		m[pieces[0]] = v
	}

	return m
}

func checkRunConfig(warningWriter io.Writer, cfg *config.Config) error {
	if cfg.Format == "json" && cfg.ShowSkipped {
		logging.Logger.Warn().Msg("show-skipped is not needed with JSON output format as that always includes them.")
	}

	if cfg.SyncUsageFile {
		missingUsageFile := make([]string, 0)
		for _, project := range cfg.Projects {
			if project.UsageFile == "" {
				missingUsageFile = append(missingUsageFile, project.Path)
			}
		}
		if len(missingUsageFile) == 1 {
			logging.Logger.Warn().Msg("Ignoring sync-usage-file as no usage-file is specified.")
		} else if len(missingUsageFile) == len(cfg.Projects) {
			logging.Logger.Warn().Msg("Ignoring sync-usage-file since no projects have a usage-file specified.")
		} else if len(missingUsageFile) > 1 {
			logging.Logger.Warn().Msgf("Ignoring sync-usage-file for following projects as no usage-file is specified for them: %s.", strings.Join(missingUsageFile, ", "))
		}
	}

	if money.GetCurrency(cfg.Currency) == nil {
		logging.Logger.Warn().Msgf("Ignoring unknown currency '%s', using USD.\n", cfg.Currency)
		cfg.Currency = "USD"
	}

	return nil
}

func (r *parallelRunner) buildRunEnv(projectContexts []*config.ProjectContext, or output.Root) map[string]interface{} {
	env := r.runCtx.EventEnvWithProjectContexts(projectContexts)

	env["runId"] = or.RunID
	env["projectCount"] = len(projectContexts)
	env["runSeconds"] = time.Now().Unix() - r.runCtx.StartTime
	env["currency"] = r.runCtx.Config.Currency

	usingCache := make([]bool, 0, len(projectContexts))
	cacheErrors := make([]string, 0, len(projectContexts))
	for _, pCtx := range projectContexts {
		usingCache = append(usingCache, pCtx.UsingCache)
		cacheErrors = append(cacheErrors, pCtx.CacheErr)
	}
	env["usingCache"] = usingCache
	env["cacheErrors"] = cacheErrors

	summary := or.FullSummary
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

	if r.pricingFetcher.MissingPricesLen() > 0 {
		env["pricesNotFoundList"] = r.pricingFetcher.MissingPricesComponents()
	}

	if n := or.ExampleProjectName(); n != "" {
		env["exampleProjectName"] = n
	}

	return env
}

func newErroredProject(provider schema.Provider, ctx *config.ProjectContext, err error) *projectOutput {
	metadata := schema.DetectProjectMetadata(ctx.ProjectConfig.Path)
	metadata.Type = "error"
	metadata.AddError(err)
	if provider != nil {
		metadata.TerraformModulePath = provider.RelativePath()
	}

	name := ctx.ProjectConfig.Name
	if name == "" {
		name = metadata.GenerateProjectName(ctx.RunContext.VCSMetadata.Remote, ctx.RunContext.IsCloudEnabled())
		if provider != nil {
			name = filepath.Join(name, provider.RelativePath())
		}
	}

	return &projectOutput{projects: []*schema.Project{schema.NewProject(name, metadata)}}
}

// runCommandFunc is a function that runs a command and returns an error this is
// used by cobra.RunE.
type runCommandFunc func(cmd *cobra.Command, args []string) error

// runCommandMiddleware is a function that wraps a runCommandFunc and returns a
// new runCommandFunc. This is used to add functionality to a command without
// modifying the command itself. Middleware can be chained together to add
// multiple pieces of functionality.
//
//nolint:deadcode,unused
type runCommandMiddleware func(ctx *config.RunContext, next runCommandFunc) runCommandFunc

// checkAPIKeyIsValid is a runCommandMiddleware that checks if the API key is
// valid before running the command.
func checkAPIKeyIsValid(ctx *config.RunContext, next runCommandFunc) runCommandFunc {
	return func(cmd *cobra.Command, args []string) error {
		if ctx.Config.APIKey == "" {
			return fmt.Errorf("%s %s %s %s %s\n%s %s.\n%s %s %s",
				ui.PrimaryString("INFRACOST_API_KEY"),
				"is not set but is required, check your",
				"environment variable is named correctly or add your API key to your",
				ui.PrimaryString(config.CredentialsFilePath()),
				"credentials file.",
				"If you recently regenerated your API key, you can retrieve it from",
				ui.LinkString(ctx.Config.DashboardEndpoint),
				"See",
				ui.LinkString("https://infracost.io/support"),
				"if you continue having issues.")
		}

		pricingClient := apiclient.NewPricingAPIClient(ctx)
		_, err := pricingClient.DoQueries([]apiclient.GraphQLQuery{
			{},
		})

		var apiError *apiclient.APIError
		if errors.As(err, &apiError) {
			if apiError.ErrorCode == apiclient.ErrorCodeAPIKeyInvalid {
				return fmt.Errorf("%s %s %s %s %s\n%s %s.\n%s %s %s",
					"Invalid API Key, please check your",
					ui.PrimaryString("INFRACOST_API_KEY"),
					"environment variable or",
					ui.PrimaryString(config.CredentialsFilePath()),
					"credentials file.",
					"If you recently regenerated your API key, you can retrieve it from",
					ui.LinkString(ctx.Config.DashboardEndpoint),
					"See",
					ui.LinkString("https://infracost.io/support"),
					"if you continue having issues.",
				)

			}
		}

		return next(cmd, args)
	}
}
