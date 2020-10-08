package aws_test

import (
	"testing"

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
					Name:            "Per Application Load Balancer",
					PriceHash:       "e31cdaab3eb4b520a8e845c058e09e75-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "aws_alb.alb1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Per Application Load Balancer",
					PriceHash:       "e31cdaab3eb4b520a8e845c058e09e75-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "aws_lb.nlb1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Per Network Load Balancer",
					PriceHash:       "cb019b908c3e3b33bb563bc3040f2e0b-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
