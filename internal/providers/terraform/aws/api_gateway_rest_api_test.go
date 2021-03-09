package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
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
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestApiGatewayRestApi_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_api_gateway_rest_api" "my_rest_api" {
			name              = "rest-api-gateway"
			description       = "Rest API Gateway"
		}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_api_gateway_rest_api.my_rest_api": map[string]interface{}{
			"monthly_requests": 21000000000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_api_gateway_rest_api.my_rest_api",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Requests (first 333M)",
					PriceHash:        "30915f094424efbda95c09ab4ee17a0b-aa6df30af0b50817c2072570cdf45eb9",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(333000000)),
				},
				{
					Name:             "Requests (next 667M)",
					PriceHash:        "30915f094424efbda95c09ab4ee17a0b-aa6df30af0b50817c2072570cdf45eb9",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(667000000)),
				},
				{
					Name:             "Requests (next 19B)",
					PriceHash:        "30915f094424efbda95c09ab4ee17a0b-aa6df30af0b50817c2072570cdf45eb9",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(19000000000)),
				},
				{
					Name:             "Requests (over 20B)",
					PriceHash:        "30915f094424efbda95c09ab4ee17a0b-aa6df30af0b50817c2072570cdf45eb9",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000000000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
