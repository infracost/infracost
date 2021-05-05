package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestRoute53Record(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_route53_zone" "zone1" {
			name = "example.com"
		}

		resource "aws_route53_record" "standard" {
			zone_id = aws_route53_zone.zone1.zone_id
			name    = "standard.example.com"
			type    = "A"
			ttl     = "300"
			records = ["10.0.0.1"]
		}
		`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "aws_route53_zone.zone1",
			SkipCheck: true,
		},
		{
			Name: "aws_route53_record.standard",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Standard queries (first 1B)",
					PriceHash:        "c07c948553cc6492cc58c7b53b8dfdf2-ce48854e53280eca3824bf5039878612",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Geo DNS queries (first 1B)",
					PriceHash:        "1565af203c9c0e9a59815a64b9c484d0-ce48854e53280eca3824bf5039878612",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Latency based routing queries (first 1B)",
					PriceHash:        "82e2ac0a19cdd4c54fea556c3f8c3892-ce48854e53280eca3824bf5039878612",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestRoute53Record_usage(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_route53_zone" "zone1" {
			name = "example.com"
		}

		resource "aws_route53_record" "my_record" {
			zone_id = aws_route53_zone.zone1.zone_id
			name    = "standard.example.com"
			type    = "A"
			ttl     = "300"
			records = ["10.0.0.1"]
		}
		`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_route53_record.my_record": map[string]interface{}{
			"monthly_standard_queries":      1100000000,
			"monthly_latency_based_queries": 1200000000,
			"monthly_geo_queries":           1500000000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "aws_route53_zone.zone1",
			SkipCheck: true,
		},
		{
			Name: "aws_route53_record.my_record",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Standard queries (first 1B)",
					PriceHash:        "c07c948553cc6492cc58c7b53b8dfdf2-ce48854e53280eca3824bf5039878612",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000000000)),
				},
				{
					Name:             "Standard queries (over 1B)",
					PriceHash:        "c07c948553cc6492cc58c7b53b8dfdf2-ce48854e53280eca3824bf5039878612",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(100000000)),
				},
				{
					Name:             "Latency based routing queries (first 1B)",
					PriceHash:        "82e2ac0a19cdd4c54fea556c3f8c3892-ce48854e53280eca3824bf5039878612",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000000000)),
				},
				{
					Name:             "Latency based routing queries (over 1B)",
					PriceHash:        "82e2ac0a19cdd4c54fea556c3f8c3892-ce48854e53280eca3824bf5039878612",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(200000000)),
				},
				{
					Name:             "Geo DNS queries (first 1B)",
					PriceHash:        "1565af203c9c0e9a59815a64b9c484d0-ce48854e53280eca3824bf5039878612",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000000000)),
				},
				{
					Name:             "Geo DNS queries (over 1B)",
					PriceHash:        "1565af203c9c0e9a59815a64b9c484d0-ce48854e53280eca3824bf5039878612",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(500000000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
