package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestOrganizationCustomRuleItem(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_config_organization_custom_rule" "my_config" {
		lambda_function_arn = "fake"
		name                = "example"
		trigger_types       = ["ConfigurationItemChangeNotification"]
	}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_config_organization_custom_rule.my_config",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Rule evaluations (first 100K)",
					PriceHash:        "b5643f5c83300f4a85d84a467af5aca4-3bf3a9bc78b9ee067586248fa8117ddb",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestOrganizationCustomRule_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_config_organization_custom_rule" "my_config" {
      lambda_function_arn = "fake"
      name                = "example"
      trigger_types       = ["ConfigurationItemChangeNotification"]
    }`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_config_organization_custom_rule.my_config": map[string]interface{}{
			"monthly_rule_evaluations": 600000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_config_organization_custom_rule.my_config",
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
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(100000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
