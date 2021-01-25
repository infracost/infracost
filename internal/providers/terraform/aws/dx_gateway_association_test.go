package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestDXGatewayAssociation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_dx_gateway_association" "gateway" {
			dx_gateway_id = "dx-123456"
			associated_gateway_id = "tgw-123456"
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_dx_gateway_association.gateway",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Transit gateway attachment",
					PriceHash:       "6dc5b2d9dbc2e95ccefef986e8fca78a-e79b72b3223a1bd297a26b680a122624",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "Data processed",
					PriceHash:       "97f8802892b5a97ef3cc31f06b4a580a-dcaa14181f6c95f2f4f3e4ccf3fee63a",
					HourlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
