package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/testutil"
)

func TestStorageBucket(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_storage_bucket" "my_storage_bucket" {
			name          = "auto-expiring-bucket"
			location      = "ASIA"
			force_destroy = true
		
			lifecycle_rule {
			condition {
				age = 3
			}
			action {
				type = "Delete"
			}
			}
		}
		`
	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_storage_bucket.my_storage_bucket",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "Data storage",
					PriceHash: "83fe976775a60e822109bcd6d2399e03-57bc5d148491a8381abaccb21ca6b4e9",
				},
			},
		},
	}
	tftest.ResourceTests(t, tf, resourceChecks)
}
