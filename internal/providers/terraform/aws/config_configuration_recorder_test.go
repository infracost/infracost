package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestConfigurationRecorderItem(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_config_configuration_recorder" "my_config" {
			name     = "example"
			role_arn = aws_iam_role.r.arn
		}
		
		resource "aws_iam_role" "r" {
			name = "awsconfig-example"
		
			assume_role_policy = <<POLICY
		{}
		POLICY
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_config_configuration_recorder.my_config",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Config items",
					PriceHash:        "8f34da0cbaaa71b45b67d99de4891d31-82a8dd965c6354fb657418947e41e612",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
				{
					Name:             "Custom config items",
					PriceHash:        "09799efb8c5c18a02b6cc1e17ab725c9-82a8dd965c6354fb657418947e41e612",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestConfigurationRecorderItem_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_config_configuration_recorder" "my_config" {
			name     = "example"
			role_arn = aws_iam_role.r.arn
		}
		
		resource "aws_iam_role" "r" {
			name = "awsconfig-example"
		
			assume_role_policy = <<POLICY
		{}
		POLICY
		}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_config_configuration_recorder.my_config": map[string]interface{}{
			"monthly_config_items":        1000,
			"monthly_custom_config_items": 2000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_config_configuration_recorder.my_config",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Config items",
					PriceHash:        "8f34da0cbaaa71b45b67d99de4891d31-82a8dd965c6354fb657418947e41e612",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000)),
				},
				{
					Name:             "Custom config items",
					PriceHash:        "09799efb8c5c18a02b6cc1e17ab725c9-82a8dd965c6354fb657418947e41e612",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(2000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
