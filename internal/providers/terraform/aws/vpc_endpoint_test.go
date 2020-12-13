package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestVpcEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_vpc_endpoint" "endpoint" {
            service_name = "com.amazonaws.region.ec2"
            vpc_id = "vpc-123456"
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_vpc_endpoint.endpoint",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "VPC endpoint",
					PriceHash:       "ef7fb85cbd68a47968dd294f49ed3517-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "Data processed",
					PriceHash:       "5fb8d7a651606fc4214684873291830f-b1ae3861dc57e2db217fa83a7420374f",
					HourlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
