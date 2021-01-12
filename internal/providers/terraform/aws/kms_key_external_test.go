package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestKMSExternalKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_kms_external_key" "kms" {}
		`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_kms_external_key.kms",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Customer master key",
					PriceHash:        "27f4c0ac50728e0b52e2eca6fae6c35b-8a6f8acec9da6fca443941d0cf1bfbef",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
