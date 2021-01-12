package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestEC2TransitGatewayVpcAttachment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_ec2_transit_gateway_vpc_attachment" "vpc_attachment" {
			subnet_ids = ["subnet-123456", "subnet-654321"]
			transit_gateway_id = "tgw-123456"
			vpc_id = "vpc-123456"
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_ec2_transit_gateway_vpc_attachment.vpc_attachment",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Transit gateway attachment",
					PriceHash:       "e7c6b648a2667e2b9f14392bc2459857-e79b72b3223a1bd297a26b680a122624",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "Data processed",
					PriceHash:       "007ede62ddee3fbb5e8ba217b6050746-dcaa14181f6c95f2f4f3e4ccf3fee63a",
					HourlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
