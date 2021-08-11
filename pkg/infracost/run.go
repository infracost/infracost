package infracost

import (
	"errors"
	"fmt"
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
	"strings"
	//log "github.com/sirupsen/logrus"
)

func runMain(runCtx *config.RunContext, diff bool) ([]byte, error) {
	projects := make([]*schema.Project, 0)
	projectContexts := make([]*config.ProjectContext, 0)

	for _, projectCfg := range runCtx.Config.Projects {
		ctx := config.NewProjectContext(runCtx, projectCfg)
		runCtx.SetCurrentProjectContext(ctx)

		//TODO: reimplement detect function
		provider, err := providers.Detect(ctx)
		if err != nil {
			m := fmt.Sprintf("%s\n\n", err)
			m += fmt.Sprintf("Use the %s flag to specify the path to one of the following:\n", ui.PrimaryString("--path"))
			m += " - Terraform plan JSON file\n - Terraform directory\n - Terraform plan file"

			if !diff {
				m += "\n - Terraform state JSON file"
			}

			return nil, clierror.NewSanitizedError(errors.New(m), "Could not detect path type")
		}
		ctx.SetContextValue("projectType", provider.Type())
		projectContexts = append(projectContexts, ctx)

		if diff && provider.Type() == "terraform_state_json" {
			m := "Cannot use Terraform state JSON with the infracost diff command.\n\n"
			m += fmt.Sprintf("Use the %s flag to specify the path to one of the following:\n", ui.PrimaryString("--path"))
			m += " - Terraform plan JSON file\n - Terraform directory\n - Terraform plan file"
			return nil, clierror.NewSanitizedError(errors.New(m), "Cannot use Terraform state JSON with the infracost diff command")
		}

		m := fmt.Sprintf("Detected %s at %s", provider.DisplayType(), ui.DisplayPath(projectCfg.Path))
		//TODO: Use gophers logger here
		log.Info(m)

		u, err := usage.LoadFromFile(projectCfg.UsageFile, runCtx.Config.SyncUsageFile)
		if err != nil {
			return nil, err
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
			return nil, err
		}

		projects = append(projects, project)

		if runCtx.Config.SyncUsageFile {
			err = usage.SyncUsageData(project, u, projectCfg.UsageFile)
			if err != nil {
				return nil, err
			}
		}

	}

	for _, project := range projects {
		if err := prices.PopulatePrices(runCtx.Config, project); err != nil {

			if e := errors.Unwrap(err); errors.Is(e, apiclient.ErrInvalidAPIKey) {
				return nil, fmt.Errorf("%v\n%s %s %s %s %s\n%s",
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
				return nil, fmt.Errorf("%v\n%s", e.Error(), "We have been notified of this issue.")
			}

			return nil, err
		}

		schema.CalculateCosts(project)
		project.CalculateDiff()
	}

	r := output.ToOutputFormat(projects)

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

	var out []byte

	switch strings.ToLower(runCtx.Config.Format) {
	case "json":
		out, err = output.ToJSON(r, opts)
	case "html":
		out, err = output.ToHTML(r, opts)
	case "diff":
		out, err = output.ToDiff(r, opts)
	default:
		out, err = output.ToTable(r, opts)
	}

	if err != nil {
		return nil, fmt.Errorf("error generating output %w", err)
	}

	return out, nil
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
