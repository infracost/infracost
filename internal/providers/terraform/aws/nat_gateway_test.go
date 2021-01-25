package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestNATGateway(t *testing.T) {
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
					Name:            "NAT gateway",
					PriceHash:       "6e137a9da0718f0ec80fb60866730ba9-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "Data processed",
					PriceHash:       "96ea6ef0b38f7b8b243f50e02dfa8fa8-b1ae3861dc57e2db217fa83a7420374f",
					HourlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestNATGateway_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_nat_gateway" "nat" {
			allocation_id = "eip-12345678"
			subnet_id     = "subnet-12345678"
		}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_nat_gateway.nat": map[string]interface{}{
			"monthly_gb_data_processed": 100,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_nat_gateway.nat",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "NAT gateway",
					PriceHash:       "6e137a9da0718f0ec80fb60866730ba9-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "Data processed",
					PriceHash:       "96ea6ef0b38f7b8b243f50e02dfa8fa8-b1ae3861dc57e2db217fa83a7420374f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
