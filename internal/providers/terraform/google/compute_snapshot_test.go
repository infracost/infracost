package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestComputeSnapshot(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
resource "google_compute_disk" "default" {
name = "test-disk"
size = 100
}

resource "google_compute_snapshot" "snapshot" {
name = "my-snapshot"
source_disk = google_compute_disk.default.name
}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "google_compute_disk.default",
			SkipCheck: true,
		},
		{
			Name: "google_compute_snapshot.snapshot",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage",
					PriceHash:        "62e9fe8c61c6e516fdf2fbb527c5ce58-b68a2427d8791aa8e79bba81a1f86e44",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestComputeSnapshot_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
resource "google_compute_disk" "default" {
name = "test-disk"
size = 100
}

resource "google_compute_snapshot" "snapshot" {
name = "my-snapshot"
source_disk = google_compute_disk.default.name
}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"google_compute_snapshot.snapshot": map[string]interface{}{
			"storage_gb": 100,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "google_compute_disk.default",
			SkipCheck: true,
		},
		{
			Name: "google_compute_snapshot.snapshot",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage",
					PriceHash:        "62e9fe8c61c6e516fdf2fbb527c5ce58-b68a2427d8791aa8e79bba81a1f86e44",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
