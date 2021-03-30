package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestVpcEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_vpc_endpoint" "interface" {
			service_name = "com.amazonaws.region.ec2"
			vpc_id = "vpc-123456"
			vpc_endpoint_type = "Interface"
		}

		resource "aws_vpc_endpoint" "gateway_loadbalancer" {
			service_name = "com.amazonaws.region.ec2"
			vpc_id = "vpc-123456"
			vpc_endpoint_type = "GatewayLoadBalancer"
		}

		resource "aws_vpc_endpoint" "multiple_interfaces" {
			service_name = "com.amazonaws.region.ec2"
			vpc_id = "vpc-123456"
			vpc_endpoint_type = "Interface"
			subnet_ids = [
				"subnet-123456",
				"subnet-654321"
			]
		}
`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_vpc_endpoint.interface",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Endpoint (Interface)",
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
		{
			Name: "aws_vpc_endpoint.gateway_loadbalancer",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Endpoint (GatewayLoadBalancer)",
					PriceHash:       "223b69fb3326be912fd0d30333e8dc50-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "Data processed",
					PriceHash:       "88513e1fd2a2e28b7ae752a813e771eb-b1ae3861dc57e2db217fa83a7420374f",
					HourlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
		{
			Name: "aws_vpc_endpoint.multiple_interfaces",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Endpoint (Interface)",
					PriceHash:       "ef7fb85cbd68a47968dd294f49ed3517-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(2)),
				},
				{
					Name:            "Data processed",
					PriceHash:       "5fb8d7a651606fc4214684873291830f-b1ae3861dc57e2db217fa83a7420374f",
					HourlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
