package apiclient

import (
	"encoding/json"
	"fmt"
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

type QueryCLISettingsResponse struct {
	CloudEnabled bool `json:"cloudEnabled"`
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
}

func NewDashboardAPIClient(ctx *config.RunContext) *DashboardAPIClient {
	return &DashboardAPIClient{
		APIClient: APIClient{
			endpoint: ctx.Config.DashboardAPIEndpoint,
			apiKey:   ctx.Config.APIKey,
			uuid:     ctx.UUID(),
		},
		shouldStoreRun: ctx.IsCloudEnabled() && !ctx.Config.IsSelfHosted(),
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

func (c *DashboardAPIClient) AddRun(ctx *config.RunContext, out output.Root) (AddRunResponse, error) {
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
		}
	}

	ctxValues := ctx.ContextValues()
	if ctx.IsInfracostComment() {
		// Clone the map to cleanup up the "command" key to show "comment".  It is
		// currently set to the sub comment (e.g. "github")
		ctxValues = make(map[string]interface{}, len(ctxValues))
		for k, v := range ctx.ContextValues() {
			ctxValues[k] = v
		}
		ctxValues["command"] = "comment"
	}

	v := map[string]interface{}{
		"run": runInput{
			ProjectResults: projectResultInputs,
			Currency:       out.Currency,
			TimeGenerated:  out.TimeGenerated.UTC(),
			Metadata:       ctxValues,
		},
	}

	q := `
	mutation($run: RunInput!) {
			addRun(run: $run) {
				id
				shareUrl
				organization {
					id
					name
				}
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
			log.Info(successMsg)
		} else {
			fmt.Fprintf(ctx.ErrWriter, "%s\n", successMsg)
		}

		response.RunID = cloudRun.Get("id").String()
		response.ShareURL = cloudRun.Get("shareUrl").String()
	}
	return response, nil
}

func (c *DashboardAPIClient) QueryCLISettings() (QueryCLISettingsResponse, error) {
	response := QueryCLISettingsResponse{}

	q := `
		query CLISettings {
        	cliSettings {
            	cloudEnabled
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
	}
	return response, nil
}
