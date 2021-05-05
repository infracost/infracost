package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestKMSCryptoKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_kms_crypto_key" "my_keys" {
			name            = "crypto-key-example"
			key_ring        = ""
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_kms_crypto_key.my_keys",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "Key versions",
					PriceHash: "48ca424cd42aff91b501178415111e68-1e9ff8daed1e44280da3d0865c8cd0c1",
				},
				{
					Name:      "Operations",
					PriceHash: "645dc979e6cfd346a9602ee5ac04f992-2e6c536b0d1e01fc280d161856794e4d",
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks, tmpDir)
}

func TestKMSCryptoKey_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "google_kms_crypto_key" "my_keys" {
		name            = "crypto-key-example"
		key_ring        = ""

		version_template {
			algorithm = "EC_SIGN_P256_SHA256"
			protection_level = "HSM"
		}
	}
	resource "google_kms_crypto_key" "with_rotate" {
		name            = "crypto-key-example"
		key_ring        = ""
		rotation_period = "100000s"
	}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"google_kms_crypto_key.my_keys": map[string]interface{}{
			"key_versions":           5000,
			"monthly_key_operations": 10000,
		},
		"google_kms_crypto_key.with_rotate": map[string]interface{}{
			"monthly_key_operations": 10000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_kms_crypto_key.my_keys",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Key versions (first 2K)",
					PriceHash:        "0d04abc08584392ef40bf656d836fbfe-a294307af4ddc1bba80ca7adc7863e3b",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(2000)),
				},
				{
					Name:             "Key versions (over 2K)",
					PriceHash:        "0d04abc08584392ef40bf656d836fbfe-fa4d023b5a11217018f841b3e948d5ef",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(3000)),
				},
				{
					Name:             "Operations",
					PriceHash:        "29038b9160b252ec14e77dbbf2babef6-2e6c536b0d1e01fc280d161856794e4d",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(10000)),
				},
			},
		},
		{
			Name: "google_kms_crypto_key.with_rotate",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Key versions",
					PriceHash:        "48ca424cd42aff91b501178415111e68-1e9ff8daed1e44280da3d0865c8cd0c1",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(2592000.0 / 100000.0)),
				},
				{
					Name:             "Operations",
					PriceHash:        "645dc979e6cfd346a9602ee5ac04f992-2e6c536b0d1e01fc280d161856794e4d",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(10000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks, tmpDir)
}
