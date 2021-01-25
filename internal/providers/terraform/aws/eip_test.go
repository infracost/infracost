package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestEIP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_eip" "eip1" {}

		resource "aws_eip" "eip2" {
			customer_owned_ipv4_pool = "pool1"
		}

		resource "aws_eip" "eip3" {
			instance = "instance1"
		}

		resource "aws_eip" "eip4" {
			network_interface = "network1"
		}
		`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_eip.eip1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "IP address (if unused)",
					PriceHash: "42572a6ef29dcca6f60464c0c0a900f7-d2c98780d7b6e36641b521f1f8145c6f",
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
