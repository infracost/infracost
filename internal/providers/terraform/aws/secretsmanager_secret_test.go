package aws_test

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAwsSecretsManagerSecretFunction(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_secretsmanager_secret" "secret" {
		name = "my-test-secret"
	}
`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_secretsmanager_secret.secret",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Secret",
					PriceHash:        "b4a5010f1c91cd8d12cc6cc367eec4fa-cb343a9216693ca16fe00f2fb3695b65",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "API requests",
					PriceHash:        "c43f680513f2f5fec806d6b1af30638a-94d92ed2c091732571fe7cdabadd7253",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestAwsSecretsManagerSecret_usage(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_secretsmanager_secret" "secret" {
		name = "my-test-secret"
	}
`
	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_secretsmanager_secret.secret": map[string]interface{}{
			"monthly_requests": 100000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_secretsmanager_secret.secret",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Secret",
					PriceHash:        "b4a5010f1c91cd8d12cc6cc367eec4fa-cb343a9216693ca16fe00f2fb3695b65",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "API requests",
					PriceHash:        "c43f680513f2f5fec806d6b1af30638a-94d92ed2c091732571fe7cdabadd7253",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
