package aws_test

import (
	"testing"

	"infracost/internal/providers/terraform/tftest"
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

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_nat_gateway.nat",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Per NAT Gateway",
					PriceHash:       "6e137a9da0718f0ec80fb60866730ba9-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "Per GB data processed",
					PriceHash:       "96ea6ef0b38f7b8b243f50e02dfa8fa8-b1ae3861dc57e2db217fa83a7420374f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
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

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_nat_gateway.nat",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Per NAT Gateway",
					PriceHash:       "6e137a9da0718f0ec80fb60866730ba9-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "Per GB data processed",
					PriceHash:       "96ea6ef0b38f7b8b243f50e02dfa8fa8-b1ae3861dc57e2db217fa83a7420374f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
