package apiclient

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	log "github.com/sirupsen/logrus"
)

type DashboardAPIClient struct {
	APIClient
	selfHostedReportingDisabled bool
}

func NewDashboardAPIClient(cfg *config.Config) *DashboardAPIClient {
	return &DashboardAPIClient{
		APIClient: APIClient{
			endpoint: cfg.DashboardAPIEndpoint,
			apiKey:   cfg.APIKey,
		},
		selfHostedReportingDisabled: cfg.IsTelemetryDisabled(),
	}
}

func (c *DashboardAPIClient) AddProjectResults(out output.Root, metadata *config.Environment) error {
	if c.selfHostedReportingDisabled {
		log.Debug("Skipping reporting project results for self-hosted Infracost")
		return nil
	}

	queries := make([]GraphQLQuery, 0, len(out.Projects))

	for _, projectResult := range out.Projects {
		v := map[string]interface{}{
			"projectResult": projectResult,
			"metadata":      metadata,
		}

		q := `
			mutation($projectResult: ProjectResultInput!, $metadata: JSONObject) {
				addProjectResult(projectResult: $projectResult, metadata: $metadata) {
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
