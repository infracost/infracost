package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestEC2TransitGatewayPeeringAttachment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_ec2_transit_gateway_peering_attachment" "peering" {
			peer_account_id = "123456789111"
			peer_region = "eu-west-1"
			peer_transit_gateway_id = "tgw-654321"
			transit_gateway_id = "tgw-123456"
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_ec2_transit_gateway_peering_attachment.peering",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Transit gateway attachment",
					PriceHash:       "3552cddb9447f1f27bc1479610a056de-e79b72b3223a1bd297a26b680a122624",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
