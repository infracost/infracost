package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestLB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_lb" "lb1" {
			load_balancer_type = "application"
		}

		resource "aws_alb" "alb1" {
		}

		resource "aws_lb" "nlb1" {
			load_balancer_type = "network"
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_lb.lb1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Application load balancer",
					PriceHash:       "e31cdaab3eb4b520a8e845c058e09e75-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Load balancer capacity units",
					PriceHash:        "5e46c73490aa808461d404c240a93a46-61920ce57954036e90af0b70fee47683",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
		{
			Name: "aws_alb.alb1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Application load balancer",
					PriceHash:       "e31cdaab3eb4b520a8e845c058e09e75-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Load balancer capacity units",
					PriceHash:        "5e46c73490aa808461d404c240a93a46-61920ce57954036e90af0b70fee47683",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
		{
			Name: "aws_lb.nlb1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Network load balancer",
					PriceHash:       "cb019b908c3e3b33bb563bc3040f2e0b-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Load balancer capacity units",
					PriceHash:        "fdf585c47c7b32f80c2290b91e06eed9-61920ce57954036e90af0b70fee47683",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
