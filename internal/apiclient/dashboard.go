package apiclient

import (
	"encoding/json"
	"time"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type DashboardAPIClient struct {
	APIClient
	selfHostedReportingDisabled bool
}

type CreateAPIKeyResponse struct {
	APIKey string `json:"apiKey"`
	Error  string `json:"error"`
}

type runInput struct {
	ProjectResults []projectResultInput   `json:"projectResultss"`
	TimeGenerated  time.Time              `json:"timeGenerated"`
	Metadata       map[string]interface{} `json:"metadata"`
}

type projectResultInput struct {
	output.Project
	Metadata map[string]interface{} `json:"metadata"`
}

func NewDashboardAPIClient(ctx *config.RunContext) *DashboardAPIClient {
	return &DashboardAPIClient{
		APIClient: APIClient{
			endpoint: ctx.Config.DashboardAPIEndpoint,
			apiKey:   ctx.Config.APIKey,
		},
		selfHostedReportingDisabled: ctx.Config.IsTelemetryDisabled(),
	}
}

func (c *DashboardAPIClient) CreateAPIKey(name string, email string) (CreateAPIKeyResponse, error) {
	d := map[string]string{"name": name, "email": email}
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

func (c *DashboardAPIClient) AddEvent(name string, env map[string]interface{}) error {
	if c.selfHostedReportingDisabled {
		log.Debug("Skipping reporting events for self-hosted Infracost")
		return nil
	}

	d := map[string]interface{}{
		"event": name,
		"env":   env,
	}

	_, err := c.doRequest("POST", "/event", d)
	return err
}

func (c *DashboardAPIClient) AddRun(ctx *config.RunContext, projectContexts []*config.ProjectContext, out output.Root) (string, error) {
	if c.selfHostedReportingDisabled {
		log.Debug("Skipping reporting project results for self-hosted Infracost")
		return "", nil
	}

	projectResultInputs := make([]projectResultInput, len(out.Projects))
	for i, project := range out.Projects {
		projectResultInputs[i] = projectResultInput{
			Project:  project,
			Metadata: projectContexts[i].ContextValues(),
		}
	}

	v := map[string]interface{}{
		"run": runInput{
			ProjectResults: projectResultInputs,
			TimeGenerated:  out.TimeGenerated,
			Metadata:       ctx.ContextValues(),
		},
	}

	q := `
	mutation($run: RunInput!) {
			addRun(run: $run) {
				id
			}
		}
	`
	results, err := c.doQueries([]GraphQLQuery{{q, v}})
	if err != nil {
		return "", err
	}

	runID := ""
	if len(results) > 0 {
		runID = results[0].Get("data.addRun.id").String()
	}
	return runID, nil
}
