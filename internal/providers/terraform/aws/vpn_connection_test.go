package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestVPNConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_vpn_connection" "vpn" {
		  customer_gateway_id = "dummy-customer-gateway-id"
		  type = "ipsec.1"
		}

		resource "aws_vpn_connection" "transit" {
		  customer_gateway_id = "dummy-customer-gateway-id"
		  type = "ipsec.1"
          transit_gateway_id = "dummy-transit-gateway-id"
		}

	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_vpn_connection.vpn",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "VPN connection",
					PriceHash:       "d6a295a59eda6edcea3cdca5a42fafde-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "aws_vpn_connection.transit",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "VPN connection",
					PriceHash:       "d6a295a59eda6edcea3cdca5a42fafde-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "Transit gateway attachment",
					PriceHash:       "06c7c9a81b26b38beacc29df55e1498b-e79b72b3223a1bd297a26b680a122624",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "Data processed",
					PriceHash:       "95f72006c31014fa4bade15b4903e2c5-dcaa14181f6c95f2f4f3e4ccf3fee63a",
					HourlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
