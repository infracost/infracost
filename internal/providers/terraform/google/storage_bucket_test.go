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
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "Network egress",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:      "Data transfer in same continent",
							PriceHash: "99caf41700f8e761f8ab246b426edbf2-8012a4febcd0213911ed09e53341a976",
						},
						{
							Name:      "Data transfer to worldwide destinations (excluding Asia & Australia) (0-1 TB)",
							PriceHash: "fa69ceb2a41a4b9cda9222f96d0e32f1-0c23081f8c5fa7d720ec507ecfd47cf6",
						},
						{
							Name:      "Data transfer to Asia Destinations (excluding China, but including Hong Kong) (0-1 TB)",
							PriceHash: "d63ba0daedaf0de514cdd32537310c00-0c23081f8c5fa7d720ec507ecfd47cf6",
						},
						{
							Name:      "Data transfer to China Destinations (excluding Hong Kong) (0-1 TB)",
							PriceHash: "237057d62af52bee885b9f353bab90e2-a62ab44470fc752864d0f5c5534f3d33",
						},
						{
							Name:      "Data transfer to Australia Destinations (0-1 TB)",
							PriceHash: "a3e569b71cd1e9d2294629e1b995c1f6-a62ab44470fc752864d0f5c5534f3d33",
						},
					},
				},
			},
		},
	}
	tftest.ResourceTests(t, tf, resourceChecks)
}
