package aws_test

import (
	"testing"

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
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
