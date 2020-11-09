package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestEKSNodeGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_eks_node_group" "example" {
		cluster_name    = "test aws_eks_node_group"
		node_group_name = "example"
		node_role_arn   = "node_role_arn"
		subnet_ids      = ["subnet_id"]
	
		scaling_config {
			desired_size = 1
			max_size     = 1
			min_size     = 1
		}
	}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_eks_node_group.example",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Linux/UNIX usage (on-demand, t3.medium)",
					PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "CPU credits",
					PriceHash:       "ccdf11d8e4c0267d78a19b6663a566c1-e8e892be2fbd1c8f42fd6761ad8977d8",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "General Purpose SSD storage (gp2)",
					PriceHash:        "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)

}
