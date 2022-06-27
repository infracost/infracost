package apiclient

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/schema"
)

type DashboardAPIClient struct {
	APIClient
	shouldStoreRun bool
}

type CreateAPIKeyResponse struct {
	APIKey string `json:"apiKey"`
	Error  string `json:"error"`
}

type AddRunResponse struct {
	RunID    string `json:"id"`
	ShareURL string `json:"shareUrl"`
}

type runInput struct {
	ProjectResults []projectResultInput   `json:"projectResults"`
	Currency       string                 `json:"currency"`
	TimeGenerated  time.Time              `json:"timeGenerated"`
	Metadata       map[string]interface{} `json:"metadata"`
}

type projectResultInput struct {
	ProjectName     string                  `json:"projectName"`
	ProjectMetadata *schema.ProjectMetadata `json:"projectMetadata"`
	PastBreakdown   *output.Breakdown       `json:"pastBreakdown"`
	Breakdown       *output.Breakdown       `json:"breakdown"`
	Diff            *output.Breakdown       `json:"diff"`
	Summary         *output.Summary         `json:"summary"`
	Metadata        map[string]interface{}  `json:"metadata"`
}

func NewDashboardAPIClient(ctx *config.RunContext) *DashboardAPIClient {
	return &DashboardAPIClient{
		APIClient: APIClient{
			endpoint: ctx.Config.DashboardAPIEndpoint,
			apiKey:   ctx.Config.APIKey,
			uuid:     ctx.UUID(),
		},
		shouldStoreRun: ctx.Config.IsCloudEnabled() && !ctx.Config.IsSelfHosted(),
	}
}

func (c *DashboardAPIClient) CreateAPIKey(name string, email string, contextVals map[string]interface{}) (CreateAPIKeyResponse, error) {
	d := map[string]interface{}{
		"name":        name,
		"email":       email,
		"os":          contextVals["os"],
		"version":     contextVals["version"],
		"fullVersion": contextVals["fullVersion"],
	}
	respBody, err := c.doRequest("POST", "/apiKeys?source=cli-register", d)
	if err != nil {
		return CreateAPIKeyResponse{}, err
	}

	var r CreateAPIKeyResponse
	err = json.Unmarshal(respBody, &r)
	if err != nil {
		return r, errors.Wrap(err, "Invalid response from API")
	}

	return r, nil
}

func (c *DashboardAPIClient) AddRun(ctx *config.RunContext, projectContexts []*config.ProjectContext, out output.Root) (AddRunResponse, error) {
	response := AddRunResponse{}

	if !c.shouldStoreRun {
		log.Debug("Skipping sending project results since it is disabled.")
		return response, nil
	}

	projectResultInputs := make([]projectResultInput, len(out.Projects))
	for i, project := range out.Projects {
		projectResultInputs[i] = projectResultInput{
			ProjectName:     project.Name,
			ProjectMetadata: project.Metadata,
			PastBreakdown:   project.PastBreakdown,
			Breakdown:       project.Breakdown,
			Diff:            project.Diff,
			Summary:         project.Summary,
			Metadata:        projectContexts[i].ContextValues(),
		}
	}

	v := map[string]interface{}{
		"run": runInput{
			ProjectResults: projectResultInputs,
			Currency:       out.Currency,
			TimeGenerated:  out.TimeGenerated.UTC(),
			Metadata:       ctx.ContextValues(),
		},
	}

	q := `
	mutation($run: RunInput!) {
			addRun(run: $run) {
				id
				shareUrl
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

		response.RunID = results[0].Get("data.addRun.id").String()
		response.ShareURL = results[0].Get("data.addRun.shareUrl").String()
	}
	return response, nil
}
