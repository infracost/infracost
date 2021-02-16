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
					Name:      "Config items",
					PriceHash: "8f34da0cbaaa71b45b67d99de4891d31-82a8dd965c6354fb657418947e41e612",
				},
				{
					Name:      "Custom config items",
					PriceHash: "09799efb8c5c18a02b6cc1e17ab725c9-82a8dd965c6354fb657418947e41e612",
				},
				{
					Name:      "Rule evaluations (first 100K)",
					PriceHash: "b5643f5c83300f4a85d84a467af5aca4-3bf3a9bc78b9ee067586248fa8117ddb",
				},
				{
					Name:      "Conformance pack evaluations (first 1M)",
					PriceHash: "59380b51cdc1fef65c1e7cf839834d1f-5809924a59f31eac9580404fc2984283",
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
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
			"monthly_config_items":                 10000,
			"monthly_custom_config_items":          10000,
			"monthly_rule_evaluations":             1000000,
			"monthly_conformance_pack_evaluations": 26000000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_config_config_rule.my_config",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Config items",
					PriceHash:        "8f34da0cbaaa71b45b67d99de4891d31-82a8dd965c6354fb657418947e41e612",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(10000)),
				},
				{
					Name:             "Custom config items",
					PriceHash:        "09799efb8c5c18a02b6cc1e17ab725c9-82a8dd965c6354fb657418947e41e612",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(10000)),
				},
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
				{
					Name:             "Conformance pack evaluations (first 1M)",
					PriceHash:        "59380b51cdc1fef65c1e7cf839834d1f-5809924a59f31eac9580404fc2984283",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000000)),
				},
				{
					Name:             "Conformance pack evaluations (next 24M)",
					PriceHash:        "59380b51cdc1fef65c1e7cf839834d1f-5809924a59f31eac9580404fc2984283",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(24000000)),
				},
				{
					Name:             "Conformance pack evaluations (over 25M)",
					PriceHash:        "59380b51cdc1fef65c1e7cf839834d1f-5809924a59f31eac9580404fc2984283",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
