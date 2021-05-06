package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestCloudFunctions(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_cloudfunctions_function" "function" {
			name        = "function-test"
			description = "My function"
			runtime     = "nodejs10"
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_cloudfunctions_function.function",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "CPU",
					PriceHash:        "49ab2e8c10b6537ef08737f3a7ae9d8d-1b88db789f176f93c9bfe3a2b0616ddd",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Memory",
					PriceHash:        "f75b0be4dd48eb627f5b50332da5dae6-54bcde0ad0d3ad3534913736ebf89cd7",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Invocations",
					PriceHash:        "fcfa5cdd12792fc691843752c9bdfb37-d1cef0b8c8b290c869ada456280c923d",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Outbound data transfer",
					PriceHash:        "7320c60393fe096cf61cb111487013a1-679daa94b986938e2c556da1f34c6f86",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}
	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestCloudFunctions_usage(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_cloudfunctions_function" "my_function" {
			name        = "function-test"
			description = "My function"
			runtime     = "nodejs10"
			available_memory_mb   = 256
		}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"google_cloudfunctions_function.my_function": map[string]interface{}{
			"request_duration_ms":          240,
			"monthly_function_invocations": 10000000,
			"monthly_outbound_data_gb":     100,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_cloudfunctions_function.my_function",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "CPU",
					PriceHash:        "49ab2e8c10b6537ef08737f3a7ae9d8d-1b88db789f176f93c9bfe3a2b0616ddd",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(10000000.0 * (400.0 / 1000.0) * 0.3)),
				},
				{
					Name:             "Memory",
					PriceHash:        "f75b0be4dd48eb627f5b50332da5dae6-54bcde0ad0d3ad3534913736ebf89cd7",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(10000000.0 * (256.0 / 1024.0) * 0.3)),
				},
				{
					Name:             "Invocations",
					PriceHash:        "fcfa5cdd12792fc691843752c9bdfb37-d1cef0b8c8b290c869ada456280c923d",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(10000000)),
				},
				{
					Name:             "Outbound data transfer",
					PriceHash:        "7320c60393fe096cf61cb111487013a1-679daa94b986938e2c556da1f34c6f86",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(100)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
