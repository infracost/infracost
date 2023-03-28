package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
)

func TestSpotInstanceRequest(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_spot_instance_request" "t3_medium" {
		ami           = "fake_ami"
		instance_type = "t3.medium"
	}

  resource "aws_spot_instance_request" "t3_large" {
    ami = "fake_ami"
    instance_type = "t3.large"
  }
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_spot_instance_request.t3_medium",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "Instance usage (Linux/UNIX, spot, t3.medium)",
					PriceHash: "c8faba8210cd512ccab6b71ca400f4de-803d7f1cd2f621429b63f791730e7935",
				},
				{
					Name:      "CPU credits",
					PriceHash: "ccdf11d8e4c0267d78a19b6663a566c1-e8e892be2fbd1c8f42fd6761ad8977d8",
				},
			},
		},
		{
			Name: "aws_spot_instance_request.t3_large",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "Instance usage (Linux/UNIX, spot, t3.large)",
					PriceHash: "3a45cd05e73384099c2ff360bdb74b74-803d7f1cd2f621429b63f791730e7935",
				},
				{
					Name:      "CPU credits",
					PriceHash: "ccdf11d8e4c0267d78a19b6663a566c1-e8e892be2fbd1c8f42fd6761ad8977d8",
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.UsageMap{}, resourceChecks)
}
