package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestApiGatewayv2Api(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
        resource "aws_apigatewayv2_api" "http" {
          name          = "test-http-api"
          protocol_type = "HTTP"
        }

        resource "aws_apigatewayv2_api" "websocket" {
          name          = "test-websocket-api"
          protocol_type = "WEBSOCKET"
        }`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_apigatewayv2_api.http",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Requests (first 300M)",
					PriceHash:        "af24853fd5a2d7b09b6c998c68aae0fb-4a9dfd3965ffcbab75845ead7a27fd47",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
			},
		},
		{
			Name: "aws_apigatewayv2_api.websocket",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Messages (first 1B)",
					PriceHash:        "a05bc87146da4c5fb7e1f26842932733-9feb253daec90eea89ff2b27827298c1",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
				{
					Name:             "Connection duration",
					PriceHash:        "7d0a09a4b1594a4ea3302640bd5ab41c-a62d9273fef0987b8d1b9a67a508acdc",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
