package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestLoggingBillingAccountSink(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_logging_billing_account_sink" "basic" {
			name            = "my-sink"
			description = "what it is"
			billing_account = "00AA00-000AAA-00AA0A" # fake
		
			destination = "storage.googleapis.com/${google_storage_bucket.bucket.name}"
		}
		
		resource "google_storage_bucket" "bucket" {
			name = "billing-logging-bucket"
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "google_storage_bucket.bucket",
			SkipCheck: true,
		},
		{
			Name: "google_logging_billing_account_sink.basic",
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

func TestLoggingBillingAccountSink_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "google_logging_billing_account_sink" "my-sink" {
		name            = "my-sink"
		description = "what it is"
		billing_account = "00AA00-000AAA-00AA0A" # fake
	
		# Can export to pubsub, cloud storage, or bigquery
		destination = "storage.googleapis.com/${google_storage_bucket.bucket.name}"
	}
	
	resource "google_storage_bucket" "bucket" {
		name = "billing-logging-bucket"
	}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"google_logging_billing_account_sink.my-sink": map[string]interface{}{
			"monthly_logging_data_gb": 100,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "google_storage_bucket.bucket",
			SkipCheck: true,
		},
		{
			Name: "google_logging_billing_account_sink.my-sink",
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
