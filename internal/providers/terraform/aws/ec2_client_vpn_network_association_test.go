package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestNewEC2ClientVPNNetworkAssociation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
          resource "aws_ec2_client_vpn_network_association" "association" {
              client_vpn_endpoint_id = "some-endpoint-id"
              subnet_id = "subnet-123456"
          }
		`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_ec2_client_vpn_network_association.association",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Endpoint association",
					PriceHash:       "5198e2e87cbf2ce28b70fdb48e9563a2-e7eda77c4cf52b2a5e814c7059c2e4c8",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
