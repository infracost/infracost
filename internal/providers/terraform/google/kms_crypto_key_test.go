package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
)

func TestKMSCryptoKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
				resource "google_kms_crypto_key" "my_keys" {
					name            = "crypto-key-example"
					key_ring        = ""
  				rotation_period = "100000s"

				
					version_template {
						algorithm = "EC_SIGN_P256_SHA256"
						protection_level = "HSM"
					}
				}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_kms_crypto_key.my_keys",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "HSM ECDSA P-256 (first 2K)",
					PriceHash: "0d04abc08584392ef40bf656d836fbfe-a294307af4ddc1bba80ca7adc7863e3b",
				},
				{
					Name:      "HSM cryptographic operations with an ECDSA P-256",
					PriceHash: "29038b9160b252ec14e77dbbf2babef6-2e6c536b0d1e01fc280d161856794e4d",
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
