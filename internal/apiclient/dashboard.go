package apiclient

import (
	"fmt"
	"strings"
	"time"

	json "github.com/json-iterator/go"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/schema"
)

type DashboardAPIClient struct {
	APIClient
}

type CreateAPIKeyResponse struct {
	APIKey string `json:"apiKey"`
	Error  string `json:"error"`
}

type AddRunResponse struct {
	RunID              string `json:"id"`
	ShareURL           string `json:"shareUrl"`
	CloudURL           string `json:"cloudUrl"`
	GovernanceFailures output.GovernanceFailures
	GovernanceComment  string             `json:"governanceComment"`
	GovernanceResults  []GovernanceResult `json:"governanceResults"`
}

type GovernanceResult struct {
	Type      string   `json:"govType"`
	Checked   int64    `json:"checked"`
	Warnings  []string `json:"warnings"`
	Failures  []string `json:"failures"`
	Unblocked []string `json:"unblocked"`
}

type QueryCLISettingsResponse struct {
	CloudEnabled       bool `json:"cloudEnabled"`
	ActualCostsEnabled bool `json:"actualCostsEnabled"`
	UsageAPIEnabled    bool `json:"usageApiEnabled"`
	TagsAPIEnabled     bool `json:"tagsApiEnabled"`
	PoliciesAPIEnabled bool `json:"policiesApiEnabled"`
}

type runInput struct {
	ProjectResults []projectResultInput   `json:"projectResults"`
	Currency       string                 `json:"currency"`
	TimeGenerated  time.Time              `json:"timeGenerated"`
	Metadata       map[string]interface{} `json:"metadata"`
	TagPolicies    []output.TagPolicy     `json:"tagPolicies,omitempty"`
	FinOpsPolicies []output.FinOpsPolicy  `json:"finopsPolicies,omitempty"`
}

type projectResultInput struct {
	ProjectName     string                  `json:"projectName"`
	ProjectMetadata *schema.ProjectMetadata `json:"projectMetadata"`
	PastBreakdown   *output.Breakdown       `json:"pastBreakdown"`
	Breakdown       *output.Breakdown       `json:"breakdown"`
	Diff            *output.Breakdown       `json:"diff"`
	Summary         *output.Summary         `json:"summary"`
}

func NewDashboardAPIClient(ctx *config.RunContext) *DashboardAPIClient {
	return &DashboardAPIClient{
		APIClient: APIClient{
			endpoint: ctx.Config.DashboardAPIEndpoint,
			apiKey:   ctx.Config.APIKey,
			uuid:     ctx.UUID(),
		},
	}
}

func newRunInput(ctx *config.RunContext, out output.Root) (*runInput, error) {
	projectResultInputs := make([]projectResultInput, len(out.Projects))
	for i, project := range out.Projects {
		projectResultInputs[i] = projectResultInput{
			ProjectName:     project.Name,
			ProjectMetadata: project.Metadata,
			PastBreakdown:   project.PastBreakdown,
			Breakdown:       project.Breakdown,
			Diff:            project.Diff,
			Summary:         project.Summary,
		}
	}

	ctxValues := ctx.ContextValues.Values()

	var metadata map[string]interface{}
	b, err := json.Marshal(out.Metadata)
	if err != nil {
		return nil, fmt.Errorf("dashboard client failed to marshal output metadata %w", err)
	}

	err = json.Unmarshal(b, &metadata)
	if err != nil {
		return nil, fmt.Errorf("dashboard client failed to unmarshal output metadata %w", err)
	}

	ctxValues["repoMetadata"] = metadata
	if ctx.IsInfracostComment() {
		ctxValues["command"] = "comment"
	}

	return &runInput{
		ProjectResults: projectResultInputs,
		Currency:       out.Currency,
		TimeGenerated:  out.TimeGenerated.UTC(),
		Metadata:       ctxValues,
		TagPolicies:    out.TagPolicies,
		FinOpsPolicies: out.FinOpsPolicies,
	}, nil
}

type CommentFormat string

var (
	CommentFormatMarkdownHTML CommentFormat = "MARKDOWN_HTML"
	CommentFormatMarkdown     CommentFormat = "MARKDOWN"
)

func (c *DashboardAPIClient) AddRun(ctx *config.RunContext, out output.Root, commentFormat CommentFormat) (AddRunResponse, error) {
	response := AddRunResponse{}

	ri, err := newRunInput(ctx, out)
	if err != nil {
		return response, err
	}

	v := map[string]interface{}{
		"run":           *ri,
		"commentFormat": commentFormat,
	}

	q := `
	mutation AddRun($run: RunInput!, $commentFormat: CommentFormat!) {
			addRun(run: $run) {
				id
				shareUrl
				cloudUrl
				organization {
					id
					name
				}

				governanceResults {
					govType
					checked
					warnings
					failures
					unblocked
				}

				governanceComment(format: $commentFormat)
			}
		}
	`
	results, err := c.doQueries([]GraphQLQuery{{q, v}})
	if err != nil {
		return response, err
	}

	if len(results) > 0 {
		if results[0].Get("errors").Exists() {
			return response, errors.New(results[0].Get("errors").String())
		}

		cloudRun := results[0].Get("data.addRun")

		orgName := cloudRun.Get("organization.name").String()
		orgMsg := ""
		if orgName != "" {
			orgMsg = fmt.Sprintf("organization '%s' in ", orgName)
		}
		successMsg := fmt.Sprintf("Estimate uploaded to %sInfracost Cloud", orgMsg)

		if ctx.Config.IsLogging() {
			log.Info().Msg(successMsg)
		} else {
			fmt.Fprintf(ctx.ErrWriter, "%s\n", successMsg)
		}

		err = json.Unmarshal([]byte(cloudRun.Raw), &response)
		if err != nil {
			return response, fmt.Errorf("failed to unmarshal addRun: %w", err)
		}

		for _, gr := range response.GovernanceResults {
			t := strings.ReplaceAll(gr.Type, "_", " ")
			if gr.Checked > 0 {
				maybePluralT := t
				if gr.Checked != 1 {
					// pluralize
					maybePluralT = strings.ReplaceAll(maybePluralT, "guardrail", "guardrails")
					maybePluralT = strings.ReplaceAll(maybePluralT, "policy", "policies")
				}
				outputGovernanceMessages(ctx, fmt.Sprintf("%d %s checked", gr.Checked, maybePluralT))
			}
			for _, msg := range gr.Unblocked {
				outputGovernanceMessages(ctx, fmt.Sprintf("%s check unblocked: %s", t, msg))
			}
			for _, msg := range gr.Warnings {
				outputGovernanceMessages(ctx, fmt.Sprintf("%s check failed: %s", t, msg))
			}
			for _, msg := range gr.Failures {
				formattedMsg := fmt.Sprintf("%s check failed: %s", t, msg)
				outputGovernanceMessages(ctx, formattedMsg)
				response.GovernanceFailures = append(response.GovernanceFailures, formattedMsg)
			}
		}
	}
	return response, nil
}

func outputGovernanceMessages(ctx *config.RunContext, msg string) {
	if ctx.Config.IsLogging() {
		log.Info().Msg(msg)
	} else {
		fmt.Fprintf(ctx.ErrWriter, "%s\n", msg)
	}
}

func (c *DashboardAPIClient) QueryCLISettings() (QueryCLISettingsResponse, error) {
	response := QueryCLISettingsResponse{}

	q := `
		query CLISettings {
        	cliSettings {
            	cloudEnabled
				actualCostsEnabled
				usageApiEnabled
				tagsApiEnabled
				policiesApiEnabled
        	}
    	}
	`
	results, err := c.doQueries([]GraphQLQuery{{q, map[string]interface{}{}}})
	if err != nil {
		return response, fmt.Errorf("query failed when requesting org settings %w", err)
	}

	if len(results) > 0 {
		if results[0].Get("errors").Exists() {
			return response, fmt.Errorf("query failed when requesting org settings, received graphql error: %s", results[0].Get("errors").String())
		}

		response.CloudEnabled = results[0].Get("data.cliSettings.cloudEnabled").Bool()
		response.ActualCostsEnabled = results[0].Get("data.cliSettings.actualCostsEnabled").Bool()
		response.UsageAPIEnabled = results[0].Get("data.cliSettings.usageApiEnabled").Bool()
		response.TagsAPIEnabled = results[0].Get("data.cliSettings.tagsApiEnabled").Bool()
		response.PoliciesAPIEnabled = results[0].Get("data.cliSettings.policiesApiEnabled").Bool()
	}
	return response, nil
}
