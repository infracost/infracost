package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
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
					Name:             "Data ingestion",
					PriceHash:        "4c00b8e26729863d2cc1f1a2d824dcf0-b1ae3861dc57e2db217fa83a7420374f",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Archival Storage",
					PriceHash:        "af1a1c7a3c3f5fc6e72de0ba26dcf55e-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Insights queries data scanned",
					PriceHash:        "e4d44a4a02daffd13cd87e63d67f30a5-b1ae3861dc57e2db217fa83a7420374f",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
