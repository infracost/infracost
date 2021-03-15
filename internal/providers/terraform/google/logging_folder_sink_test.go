package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestLoggingFolderSink(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_logging_folder_sink" "basic" {
			name   = "my-sink"
  		description = "what it is"
  		folder = google_folder.folder.name

  		destination = "storage.googleapis.com/${google_storage_bucket.bucket.name}"
		}
		
		resource "google_storage_bucket" "bucket" {
			name = "billing-logging-bucket"
		}
		
		resource "google_folder" "folder" {
			display_name = "My folder"
			parent       = "organizations/123456"
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "google_storage_bucket.bucket",
			SkipCheck: true,
		},
		{
			Name:      "google_folder.folder",
			SkipCheck: true,
		},
		{
			Name: "google_logging_folder_sink.basic",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Logging data",
					PriceHash:        "d48b116e4b03b64955766010d57bfc6b-f7a47229c228ea9f1c47bf5284626a67",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestLoggingFolderSink_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "google_logging_folder_sink" "basic" {
		name   = "my-sink"
		description = "what it is"
		folder = google_folder.folder.name

		destination = "storage.googleapis.com/${google_storage_bucket.bucket.name}"
	}
	
	resource "google_storage_bucket" "bucket" {
		name = "billing-logging-bucket"
	}
	
	resource "google_folder" "folder" {
		display_name = "My folder"
		parent       = "organizations/123456"
	}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"google_logging_folder_sink.basic": map[string]interface{}{
			"monthly_logging_data_gb": 100,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "google_storage_bucket.bucket",
			SkipCheck: true,
		},
		{
			Name:      "google_folder.folder",
			SkipCheck: true,
		},
		{
			Name: "google_logging_folder_sink.basic",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Logging data",
					PriceHash:        "d48b116e4b03b64955766010d57bfc6b-f7a47229c228ea9f1c47bf5284626a67",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(100)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
