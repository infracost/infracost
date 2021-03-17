package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestLoggingProjectFolderBucket(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_logging_project_bucket_config" "basic" {
				project    = "fake"
				location  = "global"
				retention_days = 30
				bucket_id = "_Default"
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_logging_project_bucket_config.basic",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Logging data",
					PriceHash:        "d48b116e4b03b64955766010d57bfc6b-f7a47229c228ea9f1c47bf5284626a67",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestLoggingProjectBucket_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_logging_project_bucket_config" "basic" {
			project    = "fake"
			location  = "global"
			retention_days = 30
			bucket_id = "_Default"
		}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"google_logging_project_bucket_config.basic": map[string]interface{}{
			"monthly_logging_data_gb": 100,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_logging_project_bucket_config.basic",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Logging data",
					PriceHash:        "d48b116e4b03b64955766010d57bfc6b-f7a47229c228ea9f1c47bf5284626a67",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(100)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
