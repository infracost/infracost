package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAwsSSMActivationFunction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_ssm_activation" "standard" {
			name = "test_ssm_standard_activation"
			description = "Test"
			iam_role = "my-test-iam-role"
		}
        `

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_ssm_activation.standard",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "On-Premises instance management - standard",
					SkipCheck: true,
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
