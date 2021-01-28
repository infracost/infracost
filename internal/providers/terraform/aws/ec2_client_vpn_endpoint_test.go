package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestNewEC2ClientVPNEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_ec2_client_vpn_endpoint" "endpoint" {
		  description            = "terraform-clientvpn-example"
		  server_certificate_arn = "arn:aws:acm:us-east-1:123456789123:certificate/a13e05dc-c58d-43f8-8a9b-c456f67891c2"
		  client_cidr_block = "10.0.0.0/16"

		  authentication_options {
			type = "certificate-authentication"
			root_certificate_chain_arn = "arn:aws:acm:us-east-1:123456789123:certificate/a13e05dc-c58d-43f8-8a9b-c456f67891c2"
		  }

		  connection_log_options {
			enabled               = true
			cloudwatch_log_group  = "cloudwatch-log-group"
			cloudwatch_log_stream = "cloudwatch-log-group-stream"
		  }
		}
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_ec2_client_vpn_endpoint.endpoint",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Connection",
					PriceHash:       "93f6288b5e21fd07774f34d5d18e449e-e7eda77c4cf52b2a5e814c7059c2e4c8",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
