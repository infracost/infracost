package aws_test

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAwsSSMActivationFunction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_ssm_activation" "advanced" {
			name = "test_ssm_advanced_activation"
			description = "Test"
			iam_role = "my-test-iam-role"
			registration_limit = 1001
		}
  `

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_ssm_activation.advanced",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "On-prem managed instances (advanced)",
					PriceHash:        "b6f8183d0311753e7cda0fcf60802cde-d2c98780d7b6e36641b521f1f8145c6f",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks, tmpDir)
}

func TestAwsSSMActivation_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_ssm_activation" "advanced" {
			name = "test_ssm_advanced_activation"
			description = "Test"
			iam_role = "my-test-iam-role"
		}
  `

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_ssm_activation.advanced": map[string]interface{}{
			"instance_tier": "advanced",
			"instances":     100,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_ssm_activation.advanced",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "On-prem managed instances (advanced)",
					PriceHash:        "b6f8183d0311753e7cda0fcf60802cde-d2c98780d7b6e36641b521f1f8145c6f",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(100)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks, tmpDir)
}
