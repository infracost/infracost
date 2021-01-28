package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestEKSFargateProfile_default(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_eks_cluster" "example" {
		name     = "example"
		role_arn = "arn:aws:iam::123456789012:role/Example"

		vpc_config {
			subnet_ids = ["subnet_id"]
		}
	}

	resource "aws_eks_fargate_profile" "example" {
		cluster_name           = aws_eks_cluster.example.name
		fargate_profile_name   = "example"
		pod_execution_role_arn = "arn:aws:iam::123456789012:role/Example"
		subnet_ids             = ["subnet_id"]

		selector {
			namespace = "example"
		}
	}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_eks_cluster.example",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "EKS cluster",
					PriceHash:       "f65b92131dbb1dd62d8bcb551c648398-66d0d770bee368b4f2a8f2f597eeb417",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "aws_eks_fargate_profile.example",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Per GB per hour",
					PriceHash:       "de50cdd325d12fae63879decb014633d-1fb365d8a0bc1f462690ec9d444f380c",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "Per vCPU per hour",
					PriceHash:       "7268f77ad40d22f8f8b58d5ba792d235-1fb365d8a0bc1f462690ec9d444f380c",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)

}
