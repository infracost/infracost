package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestMonitoring(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "google_monitoring_metric_descriptor" "my_monit" {
		description = "Daily sales records from all branch stores."
		display_name = "metric-descriptor"
		type = "custom.googleapis.com/stores/daily_sales"
		metric_kind = "GAUGE"
		value_type = "DOUBLE"
	}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_monitoring_metric_descriptor.my_monit",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Monitoring data (first 100K)",
					PriceHash:        "d28dec6a9e7ace87dc361cfa84d3f955-7d4951e60b67b4bc152dc716921dc3d7",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "API calls",
					PriceHash:        "ab2372dda0c91bff4b574df0f31cc087-d2f0733db2ebdc331df796416435ad3c",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestMonitoring_usage(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "google_monitoring_metric_descriptor" "my_monit" {
		description = "Daily sales records from all branch stores."
		display_name = "metric-descriptor"
		type = "custom.googleapis.com/stores/daily_sales"
		metric_kind = "GAUGE"
		value_type = "DOUBLE"
	}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"google_monitoring_metric_descriptor.my_monit": map[string]interface{}{
			"monthly_monitoring_data_mb": 500000,
			"monthly_api_calls":          1000000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_monitoring_metric_descriptor.my_monit",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Monitoring data (first 100K)",
					PriceHash:        "d28dec6a9e7ace87dc361cfa84d3f955-7d4951e60b67b4bc152dc716921dc3d7",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(100000)),
				},
				{
					Name:             "Monitoring data (next 150K)",
					PriceHash:        "d28dec6a9e7ace87dc361cfa84d3f955-47e4452af6926375f41baa01c67fafcc",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(150000)),
				},
				{
					Name:             "Monitoring data (over 250K)",
					PriceHash:        "d28dec6a9e7ace87dc361cfa84d3f955-1e895216d79057c250b99c4de8bb4889",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(250000)),
				},
				{
					Name:             "API calls",
					PriceHash:        "ab2372dda0c91bff4b574df0f31cc087-d2f0733db2ebdc331df796416435ad3c",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
