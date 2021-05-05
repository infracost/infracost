package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestConfigRuleItem(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_config_config_rule" "my_config" {
			name = "example"
		
			source {
				owner             = "AWS"
				source_identifier = ""
			}
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_config_config_rule.my_config",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Rule evaluations (first 100K)",
					PriceHash:        "b5643f5c83300f4a85d84a467af5aca4-3bf3a9bc78b9ee067586248fa8117ddb",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks, tmpDir)
}

func TestCongigRule_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_config_config_rule" "my_config" {
			name = "example"
		
			source {
				owner             = "AWS"
				source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
			}
		}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_config_config_rule.my_config": map[string]interface{}{
			"monthly_rule_evaluations": 1000000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_config_config_rule.my_config",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Rule evaluations (first 100K)",
					PriceHash:        "b5643f5c83300f4a85d84a467af5aca4-3bf3a9bc78b9ee067586248fa8117ddb",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(100000)),
				},
				{
					Name:             "Rule evaluations (next 400K)",
					PriceHash:        "b5643f5c83300f4a85d84a467af5aca4-3bf3a9bc78b9ee067586248fa8117ddb",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(400000)),
				},
				{
					Name:             "Rule evaluations (over 500K)",
					PriceHash:        "b5643f5c83300f4a85d84a467af5aca4-3bf3a9bc78b9ee067586248fa8117ddb",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(500000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks, tmpDir)
}
