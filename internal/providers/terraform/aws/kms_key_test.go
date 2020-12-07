package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestKMSKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_kms_key" "kms" {}
		`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_kms_key.kms",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Customer master key",
					PriceHash:        "27f4c0ac50728e0b52e2eca6fae6c35b-8a6f8acec9da6fca443941d0cf1bfbef",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Requests",
					PriceHash:        "7ee3fe8f4efd601e26775e1ebdf588d6-4a9dfd3965ffcbab75845ead7a27fd47",
					MonthlyCostCheck: nil,
				},
				{
					Name:             "Requests (ECC GenerateDataKeyPair)",
					PriceHash:        "b283328d4a57675972284045c9343af0-4a9dfd3965ffcbab75845ead7a27fd47",
					MonthlyCostCheck: nil,
				},
				{
					Name:             "Requests (RSA GenerateDataKeyPair)",
					PriceHash:        "b283328d4a57675972284045c9343af0-4a9dfd3965ffcbab75845ead7a27fd47",
					MonthlyCostCheck: nil,
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
