package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestCloudwatchDashboard(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
        resource "aws_cloudwatch_dashboard" "dashboard" {
          dashboard_name = "my-testing-dashboard"

          dashboard_body = <<EOF
        {
          "widgets": [
            {
              "type": "metric",
              "x": 0,
              "y": 0,
              "width": 12,
              "height": 6,
              "properties": {
                "metrics": [
                  [
                    "AWS/EC2",
                    "CPUUtilization",
                    "InstanceId",
                    "i-012345"
                  ]
                ],
                "period": 300,
                "stat": "Average",
                "region": "us-east-1",
                "title": "EC2 Instance CPU"
              }
            },
            {
              "type": "text",
              "x": 0,
              "y": 7,
              "width": 3,
              "height": 3,
              "properties": {
                "markdown": "Hello world"
              }
            }
          ]
        }
        EOF
        }
`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_cloudwatch_dashboard.dashboard",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Dashboard",
					PriceHash:        "0d9249c99a5605c643f4505314d483f7-e5d57ccffe964e509c0afb79efbe1987",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
