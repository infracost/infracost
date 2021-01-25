package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
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
					Name:             "ECC GenerateDataKeyPair requests",
					PriceHash:        "b283328d4a57675972284045c9343af0-4a9dfd3965ffcbab75845ead7a27fd47",
					MonthlyCostCheck: nil,
				},
				{
					Name:             "RSA GenerateDataKeyPair requests",
					PriceHash:        "b283328d4a57675972284045c9343af0-4a9dfd3965ffcbab75845ead7a27fd47",
					MonthlyCostCheck: nil,
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestKMSKey_RSA_2048(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_kms_key" "kms" {
			customer_master_key_spec = "RSA_2048"
		}
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
					Name:             "Requests (RSA 2048)",
					PriceHash:        "cf7b71f1cff51c5b2963048440c65ddc-4a9dfd3965ffcbab75845ead7a27fd47",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestKMSKey_Asymmetric(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_kms_key" "kms" {
			customer_master_key_spec = "RSA_3072"
		}
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
					Name:             "Requests (asymmetric)",
					PriceHash:        "e6c7bc01771a8886348e2727083eab1b-4a9dfd3965ffcbab75845ead7a27fd47",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
