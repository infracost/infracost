package aws_test

import (
	"testing"

	"infracost/internal/providers/terraform/tftest"
	"infracost/pkg/schema"
	"infracost/pkg/testutil"

	"github.com/shopspring/decimal"
)

func TestAwsNatGateway(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_nat_gateway" "nat" {
			allocation_id = "eip-12345678"
			subnet_id     = "subnet-12345678"
		}`

	resources, err := tftest.RunCostCalculation(tf)
	if err != nil {
		t.Error(err)
	}

	expectedPriceHashes := [][]string{
		{"aws_nat_gateway.nat", "Per NAT Gateway", "6e137a9da0718f0ec80fb60866730ba9-d2c98780d7b6e36641b521f1f8145c6f"},
		{"aws_nat_gateway.nat", "Per GB data processed", "96ea6ef0b38f7b8b243f50e02dfa8fa8-b1ae3861dc57e2db217fa83a7420374f"},
	}
	testutil.CheckPriceHashes(t, resources, expectedPriceHashes)

	var costComponent *schema.CostComponent

	costComponent = testutil.FindCostComponent(resources, "aws_nat_gateway.nat", "Per NAT Gateway")
	testutil.CheckCost(t, "aws_nat_gateway.nat", costComponent, "hourly", costComponent.Price())

	costComponent = testutil.FindCostComponent(resources, "aws_nat_gateway.nat", "Per GB data processed")
	testutil.CheckCost(t, "aws_nat_gateway.nat", costComponent, "monthly", decimal.Zero)
}

func TestAwsNatGateway_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_nat_gateway" "nat" {
			allocation_id = "eip-12345678"
			subnet_id     = "subnet-12345678"
		}
		
		resource "infracost_aws_nat_gateway" "nat" {
			resources = [aws_nat_gateway.nat.id]

			gb_data_processed_monthly {
				value = 100
			}
		}
		`

	resources, err := tftest.RunCostCalculation(tf)
	if err != nil {
		t.Error(err)
	}

	costComponent := testutil.FindCostComponent(resources, "aws_nat_gateway.nat", "Per GB data processed")
	testutil.CheckCost(t, "aws_nat_gateway.nat", costComponent, "monthly", costComponent.Price().Mul(decimal.NewFromInt(100)))
}
