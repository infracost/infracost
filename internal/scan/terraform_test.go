package scan_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/providers/terraform"
	"github.com/infracost/infracost/internal/scan"
	"github.com/infracost/infracost/internal/schema"
)

func TestTerraformPlanScanner_ScanPlan(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/policy" {
			_, _ = w.Write([]byte(`
			{
				"result": [
					{
						"id": "aws_instance.gp3",
						"title": "Upgrade EC2 gp2 volumes to gp3.",
						"description": "Upgrade undefined root block device volume type to gp3",
						"address": "aws_instance.web_app",
						"resource_type": "Instance",
						"resource_attributes": {
						  "RootBlockDevice": {
							"Type": "gp3"
						  }
						},
						"suggested": "gp3"						
					}
				]
			}`))
			return
		}

		w.WriteHeader(500)
	}))

	runCtx := &config.RunContext{
		Config: &config.Config{
			PolicyAPIEndpoint: s.URL,
		},
	}

	baseCost := decimal.NewFromInt(15)
	newCost := decimal.NewFromInt(5)

	var called int
	ps := scan.NewTerraformPlanScanner(runCtx, newDiscardLogger(), func(ctx *config.RunContext, c *apiclient.PricingAPIClient, r *schema.Resource) error {
		t.Helper()

		if called == 0 {
			component := r.SubResources[0].CostComponents[0]
			assert.Contains(t, component.Name, "gp2")
			component.SetPrice(baseCost)
		}

		if called == 1 {
			component := r.SubResources[0].CostComponents[0]
			assert.Contains(t, component.Name, "gp3")
			component.SetPrice(newCost)
		}

		if called > 1 {
			t.Errorf("unexpect call to get prices for project")
		}

		called += 1
		return nil
	})

	ctx := config.NewProjectContext(&config.RunContext{Config: &config.Config{}}, &config.Project{Path: "./testdata/simple_project"}, logrus.Fields{})
	hclp, err := terraform.NewHCLProvider(
		ctx,
		nil,
	)
	require.NoError(t, err)

	projects, err := hclp.LoadResources(map[string]*schema.UsageData{})
	require.NoError(t, err)
	plans, err := hclp.LoadPlanJSONs()
	require.NoError(t, err)

	project := projects[0]
	err = ps.ScanPlan(project, plans[0].JSON)
	require.NoError(t, err)

	assert.Len(t, project.Metadata.Policies, 1)
	b, err := json.Marshal(project.Metadata.Policies[0])
	str := string(b)

	require.NoError(t, err)
	assert.JSONEq(t, `
{
  "id": "aws_instance.gp3",
  "title": "Upgrade EC2 gp2 volumes to gp3.",
  "description": "Upgrade undefined root block device volume type to gp3",
  "resource_type": "Instance",
  "resource_attributes": {
    "RootBlockDevice": {
      "Type": "gp3"
    }
  },
  "address": "aws_instance.web_app",
  "suggested": "gp3",
  "no_cost": false,
  "cost": "500"
}
`, str)
}

func newDiscardLogger() *logrus.Entry {
	l := logrus.New()
	l.SetOutput(io.Discard)
	return l.WithFields(logrus.Fields{})
}
