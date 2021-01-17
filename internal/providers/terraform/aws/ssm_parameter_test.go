package aws_test

import (
	"github.com/shopspring/decimal"
	"testing"

	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAwsSSMParameterFunction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
        resource "aws_ssm_parameter" "advanced" {
            name = "my-advanced-ssm-parameter"
			type = "String"
			value = "Advanced Parameter"
			tier = "Advanced"
        }

		resource "aws_ssm_parameter" "standard" {
			name = "my-standard-ssm-parameter"
			type = "String"
			value = "Standard Parameter"
		}
        `

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_ssm_parameter.advanced",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Parameter storage - advanced",
					PriceHash:        "d5db437b8b7a6df9c701534aefab452b-1065e83bbc0d4959dda26a1848f3e3eb",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "API interactions - advanced",
					PriceHash:        "8f8b82df990877781864a9489c71fd99-7c35c68819b19a7ff1d898cc5a198a7f",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "aws_ssm_parameter.standard",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "Parameter storage - standard",
					SkipCheck: true,
				},
				{
					Name:      "API interactions - standard",
					SkipCheck: true,
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
