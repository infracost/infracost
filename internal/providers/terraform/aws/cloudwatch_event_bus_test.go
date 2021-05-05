package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestCloudwatchEventBus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_cloudwatch_event_bus" "my_events" {
			name = "chat-messages"
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_cloudwatch_event_bus.my_events",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Custom events published",
					PriceHash:        "77f325b5d616e704eb3f2a1eb928db6b-8c1e6098ebbcb0309f3b80ec6b497ddc",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Third-party events published",
					PriceHash:        "2b3bfcb7e2b290419a5b9feb11c73693-8c1e6098ebbcb0309f3b80ec6b497ddc",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Archive processing",
					PriceHash:        "d4550034186eb4de0292de23a4e8cd6e-dcaa14181f6c95f2f4f3e4ccf3fee63a",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Archive storage",
					PriceHash:        "866ca0470bc4656cedff737e0d766e07-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Schema discovery",
					PriceHash:        "0de4fc235bdb75f4dd95f9ca253d38c9-62b04c38def877db6fc9e4409fdfb4a7",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks, tmpDir)
}

func TestCloudwatchEventBus_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `resource "aws_cloudwatch_event_bus" "my_events" {
		name = "chat-messages"
	}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_cloudwatch_event_bus.my_events": map[string]interface{}{
			"monthly_custom_events":           1000000,
			"monthly_third_party_events":      2000000,
			"monthly_archive_processing_gb":   100,
			"archive_storage_gb":              200,
			"monthly_schema_discovery_events": 1000000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_cloudwatch_event_bus.my_events",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Custom events published",
					PriceHash:        "77f325b5d616e704eb3f2a1eb928db6b-8c1e6098ebbcb0309f3b80ec6b497ddc",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000000)),
				},
				{
					Name:             "Third-party events published",
					PriceHash:        "2b3bfcb7e2b290419a5b9feb11c73693-8c1e6098ebbcb0309f3b80ec6b497ddc",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(2000000)),
				},
				{
					Name:             "Archive processing",
					PriceHash:        "d4550034186eb4de0292de23a4e8cd6e-dcaa14181f6c95f2f4f3e4ccf3fee63a",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(100)),
				},
				{
					Name:             "Archive storage",
					PriceHash:        "866ca0470bc4656cedff737e0d766e07-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(200)),
				},
				{
					Name:             "Schema discovery",
					PriceHash:        "0de4fc235bdb75f4dd95f9ca253d38c9-62b04c38def877db6fc9e4409fdfb4a7",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks, tmpDir)
}
