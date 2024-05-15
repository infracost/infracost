package aws_test

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestEKSNodeGroupGoldenFile(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tftest.GoldenFileHCLResourceTestsWithOpts(t, "eks_node_group_test", tftest.DefaultGoldenFileOptions(), func(ctx *config.RunContext) {
		ctx.Config.GraphEvaluator = true
	})
}

func TestEKSNodeGroup_spot(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_eks_node_group" "example" {
		cluster_name    = "test_aws_eks_node_group"
		node_group_name = "example"
		node_role_arn   = "node_role_arn"
		subnet_ids      = ["subnet_id"]
		capacity_type   = "SPOT"

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
					Name:            "Instance usage (Linux/UNIX, spot, t3.medium)",
					PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-803d7f1cd2f621429b63f791730e7935",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "CPU credits",
					PriceHash:        "ccdf11d8e4c0267d78a19b6663a566c1-e8e892be2fbd1c8f42fd6761ad8977d8",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
				{
					Name:             "Storage (general purpose SSD, gp2)",
					PriceHash:        "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.UsageMap{}, resourceChecks)
}
