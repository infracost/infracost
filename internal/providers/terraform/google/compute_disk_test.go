package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestComputeDisk(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_compute_disk" "standard_default" {
			name = "standard_default"
			type = "pd-standard"
		}

		resource "google_compute_disk" "ssd_default" {
			name = "ssd_default"
			type = "pd-ssd"
		}

		resource "google_compute_disk" "size" {
			name = "size"
			type = "pd-standard"
			size = 20
		}

		resource "google_compute_image" "image_disk_size" {
			name = "image_disk_size"
			disk_size_gb = 30
		}

		resource "google_compute_disk" "image_disk_size" {
			name = "image_disk_size"
			type = "pd-standard"
			image = google_compute_image.image_disk_size.self_link
		}

		resource "google_compute_image" "image_source_image" {
			name = "image_source_image"
			source_image = google_compute_image.image_disk_size.self_link
		}

		resource "google_compute_disk" "image_source_image" {
			name = "image_source_image"
			type = "pd-standard"
			image = google_compute_image.image_source_image.self_link
		}

		resource "google_compute_snapshot" "snapshot_source_disk" {
			name = "snapshot_source_disk"
			source_disk = google_compute_disk.size.name
		}

		resource "google_compute_image" "image_source_snapshot" {
			name = "image_source_snapshot"
			source_snapshot = google_compute_snapshot.snapshot_source_disk.self_link
		}

		resource "google_compute_disk" "image_source_snapshot" {
			name = "image_source_snapshot"
			type = "pd-standard"
			image = google_compute_image.image_source_snapshot.self_link
		}

		resource "google_compute_disk" "snapshot_source_disk" {
			name = "snapshot_source_disk"
			type = "pd-standard"
			snapshot = google_compute_snapshot.snapshot_source_disk.self_link
		}
		`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_compute_disk.standard_default",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Standard provisioned storage (pd-standard)",
					PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(500)),
				},
			},
		},
		{
			Name: "google_compute_disk.ssd_default",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "SSD provisioned storage (pd-ssd)",
					PriceHash:        "7317191236b3f20b4e8122bddb65e5cf-57bc5d148491a8381abaccb21ca6b4e9",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100)),
				},
			},
		},
		{
			Name: "google_compute_disk.size",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Standard provisioned storage (pd-standard)",
					PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
			},
		},
		{
			Name:      "google_compute_image.image_disk_size",
			SkipCheck: true,
		},
		{
			Name: "google_compute_disk.image_disk_size",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Standard provisioned storage (pd-standard)",
					PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(30)),
				},
			},
		},
		{
			Name:      "google_compute_image.image_source_image",
			SkipCheck: true,
		},
		{
			Name: "google_compute_disk.image_source_image",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Standard provisioned storage (pd-standard)",
					PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(30)),
				},
			},
		},
		{
			Name:      "google_compute_snapshot.snapshot_source_disk",
			SkipCheck: true,
		},
		{
			Name:      "google_compute_image.image_source_snapshot",
			SkipCheck: true,
		},
		{
			Name: "google_compute_disk.image_source_snapshot",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Standard provisioned storage (pd-standard)",
					PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
			},
		},
		{
			Name: "google_compute_disk.snapshot_source_disk",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Standard provisioned storage (pd-standard)",
					PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
