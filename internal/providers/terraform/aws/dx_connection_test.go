package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestDXConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_dx_connection" "my_dx_connection" {
  			bandwidth = "1Gbps"
  			location = "EqDC2"
  			name = "Test"
		}
`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_dx_connection.my_dx_connection": map[string]interface{}{
			"monthly_outbound_region_to_dx_location_gb": 100,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_dx_connection.my_dx_connection",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "DX connection",
					PriceHash:       "a4059a0e409557d04c555e419764e885-1fb365d8a0bc1f462690ec9d444f380c",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "Outbound data transfer to dx location EqDC2",
					PriceHash:       "e4ca252b0e2ad70c7e9c5138ade2eded-b1ae3861dc57e2db217fa83a7420374f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
