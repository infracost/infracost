package aws_test

import (
	"github.com/shopspring/decimal"
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestRoute53ResolverEndpointFunction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_route53_resolver_endpoint" "test" {
			direction = "INBOUND"
			security_group_ids = ["sg-1233456"]

			ip_address {
				subnet_id = "subnet-123456"
			}

			ip_address {
				subnet_id = "subnet-654321"
			}
		}
`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_route53_resolver_endpoint.test",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Resolver endpoints",
					PriceHash:       "82cdab229057bbe4c4558d1496adccc7-66d0d770bee368b4f2a8f2f597eeb417",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(2)),
				},
				{
					Name:             "DNS queries (first 1B)",
					PriceHash:        "417aa83709d67da384d1770246bbce48-ce48854e53280eca3824bf5039878612",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks, tmpDir)
}

func TestRoute53ResolverEndpointUsage1B(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_route53_resolver_endpoint" "test" {
			direction = "INBOUND"
			security_group_ids = ["sg-1233456"]

			ip_address {
				subnet_id = "subnet-123456"
			}

			ip_address {
				subnet_id = "subnet-654321"
			}
		}
`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_route53_resolver_endpoint.test": map[string]interface{}{
			"monthly_queries": 100000000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_route53_resolver_endpoint.test",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Resolver endpoints",
					PriceHash:       "82cdab229057bbe4c4558d1496adccc7-66d0d770bee368b4f2a8f2f597eeb417",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(2)),
				},
				{
					Name:             "DNS queries (first 1B)",
					PriceHash:        "417aa83709d67da384d1770246bbce48-ce48854e53280eca3824bf5039878612",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100000000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks, tmpDir)
}

func TestRoute53ResolverEndpointUsage2B(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_route53_resolver_endpoint" "test" {
			direction = "INBOUND"
			security_group_ids = ["sg-1233456"]

			ip_address {
				subnet_id = "subnet-123456"
			}

			ip_address {
				subnet_id = "subnet-654321"
			}
		}
`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_route53_resolver_endpoint.test": map[string]interface{}{
			"monthly_queries": 2000000000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_route53_resolver_endpoint.test",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Resolver endpoints",
					PriceHash:       "82cdab229057bbe4c4558d1496adccc7-66d0d770bee368b4f2a8f2f597eeb417",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(2)),
				},
				{
					Name:             "DNS queries (first 1B)",
					PriceHash:        "417aa83709d67da384d1770246bbce48-ce48854e53280eca3824bf5039878612",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1000000000)),
				},
				{
					Name:             "DNS queries (over 1B)",
					PriceHash:        "417aa83709d67da384d1770246bbce48-ce48854e53280eca3824bf5039878612",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1000000000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks, tmpDir)
}
