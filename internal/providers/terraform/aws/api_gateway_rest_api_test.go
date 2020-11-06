package aws_test

import (
	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
	"testing"
)

func TestApiGatewayRestApi(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
        resource "aws_api_gateway_rest_api" "api" {
          name              = "rest-api-gateway"
          description       = "Rest API Gateway"  
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_api_gateway_rest_api.api",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Requests (first 333M)",
					PriceHash:        "30915f094424efbda95c09ab4ee17a0b-aa6df30af0b50817c2072570cdf45eb9",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
				{
					Name:             "Requests (next 667M)",
					PriceHash:        "30915f094424efbda95c09ab4ee17a0b-aa6df30af0b50817c2072570cdf45eb9",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
				{
					Name:             "Requests (next 19B)",
					PriceHash:        "30915f094424efbda95c09ab4ee17a0b-aa6df30af0b50817c2072570cdf45eb9",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
				{
					Name:             "Requests (Over 20B)",
					PriceHash:        "30915f094424efbda95c09ab4ee17a0b-aa6df30af0b50817c2072570cdf45eb9",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
