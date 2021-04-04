package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestComputeImage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "google_compute_image" "empty" {
		name = "example-image"
	}
	
	resource "google_compute_disk" "disk" {
		name  = "test-disk"
		size = 1000
	}

	resource "google_compute_image" "image" {
		name = "image_source_image"
		disk_size_gb = 100
	}

	resource "google_compute_snapshot" "snapshot" {
		name = "snapshot_source_disk"
		source_disk = google_compute_disk.disk.self_link
	}

	resource "google_compute_image" "with_disk_size" {
		name = "example-image"
		disk_size_gb = 500
	}
	
	resource "google_compute_image" "with_source_disk" {
		name = "example-image"
		source_disk = google_compute_disk.disk.self_link
	}

	resource "google_compute_image" "with_source_image" {
		name = "example-image"
		source_image = google_compute_image.image.self_link
	}

	resource "google_compute_image" "with_source_snapshot" {
		name = "example-image"
		source_snapshot = google_compute_snapshot.snapshot.self_link
	}
	
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "google_compute_disk.disk",
			SkipCheck: true,
		},
		{
			Name:      "google_compute_image.image",
			SkipCheck: true,
		},
		{
			Name:      "google_compute_snapshot.snapshot",
			SkipCheck: true,
		},
		{
			Name: "google_compute_image.empty",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage",
					PriceHash:        "1ab5c2278934b99f51df9f4ed5f3a1ec-57bc5d148491a8381abaccb21ca6b4e9",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
		{
			Name: "google_compute_image.with_disk_size",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage",
					PriceHash:        "1ab5c2278934b99f51df9f4ed5f3a1ec-57bc5d148491a8381abaccb21ca6b4e9",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(500)),
				},
			},
		},
		{
			Name: "google_compute_image.with_source_disk",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage",
					PriceHash:        "1ab5c2278934b99f51df9f4ed5f3a1ec-57bc5d148491a8381abaccb21ca6b4e9",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1000)),
				},
			},
		},
		{
			Name: "google_compute_image.with_source_image",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage",
					PriceHash:        "1ab5c2278934b99f51df9f4ed5f3a1ec-57bc5d148491a8381abaccb21ca6b4e9",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100)),
				},
			},
		},
		{
			Name: "google_compute_image.with_source_snapshot",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage",
					PriceHash:        "1ab5c2278934b99f51df9f4ed5f3a1ec-57bc5d148491a8381abaccb21ca6b4e9",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestComputeImage_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "google_compute_image" "empty" {
		name = "example-image"
	}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"google_compute_image.empty": map[string]interface{}{
			"storage_gb": 100,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_compute_image.empty",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage",
					PriceHash:        "1ab5c2278934b99f51df9f4ed5f3a1ec-57bc5d148491a8381abaccb21ca6b4e9",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
