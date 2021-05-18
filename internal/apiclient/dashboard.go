package apiclient

import (
	"encoding/json"

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

func (c *DashboardAPIClient) AddProjectResults(projectContexts []*config.ProjectContext, out output.Root) error {
	if c.selfHostedReportingDisabled {
		log.Debug("Skipping reporting project results for self-hosted Infracost")
		return nil
	}

	queries := make([]GraphQLQuery, 0, len(out.Projects))

	for i, projectResult := range out.Projects {
		projectCtx := projectContexts[i]

		v := map[string]interface{}{
			"projectResult": projectResult,
			"env":           projectCtx.AllContextValues(),
		}

		q := `
			mutation($projectResult: ProjectResultInput!, $env: JSONObject) {
				addProjectResult(projectResult: $projectResult, env: $env) {
					id
				}
			}
		`

		queries = append(queries, GraphQLQuery{q, v})
	}

	_, err := c.doQueries(queries)
	if err != nil {
		return err
	}

	return nil
}
