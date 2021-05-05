package aws_test

import (
	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
	"testing"
)

func TestGetRoute53HealthCheckAWS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_route53_health_check" "simple" {
  			failure_threshold = "5"
  			fqdn              = "example.com"
  			port              = 80
  			request_interval  = "30"
  			resource_path     = "/"
  			type              = "HTTP"
		}

		resource "aws_route53_health_check" "https" {
  			failure_threshold = "5"
  			fqdn              = "example.com"
  			port              = 443
  			request_interval  = "30"
  			resource_path     = "/"
  			type              = "HTTPS"
		}

		resource "aws_route53_health_check" "ssl_string_match" {
  			failure_threshold = "5"
  			fqdn              = "example.com"
  			port              = 443
  			request_interval  = "30"
  			resource_path     = "/"
  			type              = "HTTPS_STR_MATCH"
			search_string     = "TestingTesting"
		}

		resource "aws_route53_health_check" "ssl_latency_string_match" {
  			failure_threshold = "5"
  			fqdn              = "example.com"
  			port              = 443
  			request_interval  = "30"
  			resource_path     = "/"
  			type              = "HTTPS_STR_MATCH"
			search_string     = "TestingTesting"
			measure_latency   = true
		}

		resource "aws_route53_health_check" "ssl_latency_interval_string_match" {
  			failure_threshold = "5"
  			fqdn              = "example.com"
  			port              = 443
  			request_interval  = "10"
  			resource_path     = "/"
  			type              = "HTTPS_STR_MATCH"
			search_string     = "TestingTesting"
			measure_latency   = true
		}
`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_route53_health_check.simple",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Health check",
					PriceHash:        "ce31281484f92e4e162552798229e7f1-a9191d0a7972a4ac9c0e44b9ea6310bb",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "aws_route53_health_check.https",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Health check",
					PriceHash:        "ce31281484f92e4e162552798229e7f1-a9191d0a7972a4ac9c0e44b9ea6310bb",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Optional features",
					PriceHash:        "2f754f0d7b985fea079c82627f121bbb-a9191d0a7972a4ac9c0e44b9ea6310bb",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "aws_route53_health_check.ssl_string_match",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Health check",
					PriceHash:        "ce31281484f92e4e162552798229e7f1-a9191d0a7972a4ac9c0e44b9ea6310bb",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Optional features",
					PriceHash:        "2f754f0d7b985fea079c82627f121bbb-a9191d0a7972a4ac9c0e44b9ea6310bb",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(2)),
				},
			},
		},
		{
			Name: "aws_route53_health_check.ssl_latency_string_match",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Health check",
					PriceHash:        "ce31281484f92e4e162552798229e7f1-a9191d0a7972a4ac9c0e44b9ea6310bb",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Optional features",
					PriceHash:        "2f754f0d7b985fea079c82627f121bbb-a9191d0a7972a4ac9c0e44b9ea6310bb",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(3)),
				},
			},
		},
		{
			Name: "aws_route53_health_check.ssl_latency_interval_string_match",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Health check",
					PriceHash:        "ce31281484f92e4e162552798229e7f1-a9191d0a7972a4ac9c0e44b9ea6310bb",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Optional features",
					PriceHash:        "2f754f0d7b985fea079c82627f121bbb-a9191d0a7972a4ac9c0e44b9ea6310bb",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(4)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks, tmpDir)
}

func TestGetRoute53HealthCheckOutsideAWS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_route53_health_check" "simple" {
  			failure_threshold = "5"
  			fqdn              = "example.com"
  			port              = 80
  			request_interval  = "30"
  			resource_path     = "/"
  			type              = "HTTP"
		}

		resource "aws_route53_health_check" "https" {
  			failure_threshold = "5"
  			fqdn              = "example.com"
  			port              = 443
  			request_interval  = "30"
  			resource_path     = "/"
  			type              = "HTTPS"
		}

		resource "aws_route53_health_check" "ssl_string_match" {
  			failure_threshold = "5"
  			fqdn              = "example.com"
  			port              = 443
  			request_interval  = "30"
  			resource_path     = "/"
  			type              = "HTTPS_STR_MATCH"
			search_string     = "TestingTesting"
		}

		resource "aws_route53_health_check" "ssl_latency_string_match" {
  			failure_threshold = "5"
  			fqdn              = "example.com"
  			port              = 443
  			request_interval  = "30"
  			resource_path     = "/"
  			type              = "HTTPS_STR_MATCH"
			search_string     = "TestingTesting"
			measure_latency   = true
		}

		resource "aws_route53_health_check" "ssl_latency_interval_string_match" {
  			failure_threshold = "5"
  			fqdn              = "example.com"
  			port              = 443
  			request_interval  = "10"
  			resource_path     = "/"
  			type              = "HTTPS_STR_MATCH"
			search_string     = "TestingTesting"
			measure_latency   = true
		}
`
	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_route53_health_check.simple": map[string]interface{}{
			"endpoint_type": "non_aws",
		},
		"aws_route53_health_check.https": map[string]interface{}{
			"endpoint_type": "non_aws",
		},
		"aws_route53_health_check.ssl_string_match": map[string]interface{}{
			"endpoint_type": "Non_AWS",
		},
		"aws_route53_health_check.ssl_latency_string_match": map[string]interface{}{
			"endpoint_type": "Non_AWS",
		},
		"aws_route53_health_check.ssl_latency_interval_string_match": map[string]interface{}{
			"endpoint_type": "Non_AWS",
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_route53_health_check.simple",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Health check",
					PriceHash:        "ffb695281d2d6aa969f467655260a977-a9191d0a7972a4ac9c0e44b9ea6310bb",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "aws_route53_health_check.https",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Health check",
					PriceHash:        "ffb695281d2d6aa969f467655260a977-a9191d0a7972a4ac9c0e44b9ea6310bb",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Optional features",
					PriceHash:        "87ab632b0f2388dee4f62036923ec5c5-a9191d0a7972a4ac9c0e44b9ea6310bb",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "aws_route53_health_check.ssl_string_match",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Health check",
					PriceHash:        "ffb695281d2d6aa969f467655260a977-a9191d0a7972a4ac9c0e44b9ea6310bb",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Optional features",
					PriceHash:        "87ab632b0f2388dee4f62036923ec5c5-a9191d0a7972a4ac9c0e44b9ea6310bb",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(2)),
				},
			},
		},
		{
			Name: "aws_route53_health_check.ssl_latency_string_match",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Health check",
					PriceHash:        "ffb695281d2d6aa969f467655260a977-a9191d0a7972a4ac9c0e44b9ea6310bb",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Optional features",
					PriceHash:        "87ab632b0f2388dee4f62036923ec5c5-a9191d0a7972a4ac9c0e44b9ea6310bb",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(3)),
				},
			},
		},
		{
			Name: "aws_route53_health_check.ssl_latency_interval_string_match",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Health check",
					PriceHash:        "ffb695281d2d6aa969f467655260a977-a9191d0a7972a4ac9c0e44b9ea6310bb",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Optional features",
					PriceHash:        "87ab632b0f2388dee4f62036923ec5c5-a9191d0a7972a4ac9c0e44b9ea6310bb",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(4)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks, tmpDir)
}
