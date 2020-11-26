package aws_test

import (
	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
	"testing"
)

func TestCloudwatchLogGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
        resource "aws_cloudwatch_log_group" "logs" {
          name              = "log-group"
        }`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_cloudwatch_log_group.logs",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Collect (Data Ingestion)",
					PriceHash:        "4c00b8e26729863d2cc1f1a2d824dcf0-b1ae3861dc57e2db217fa83a7420374f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
				{
					Name:             "Store (Archival)",
					PriceHash:        "af1a1c7a3c3f5fc6e72de0ba26dcf55e-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
