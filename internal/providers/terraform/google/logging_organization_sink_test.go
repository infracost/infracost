package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestLoggingOrgSink(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_logging_organization_sink" "basic" {
			name   = "basic"
			description = "what it is"
			org_id = "123456789"
		
			destination = "storage.googleapis.com/${google_storage_bucket.log-bucket.name}"
		}
		
		resource "google_storage_bucket" "log-bucket" {
			name = "organization-logging-bucket"
		}
		
		resource "google_project_iam_member" "log-writer" {
			role = "roles/storage.objectCreator"
		
			member = google_logging_organization_sink.basic.writer_identity
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "google_storage_bucket.log-bucket",
			SkipCheck: true,
		},
		{
			Name:      "google_project_iam_member.log-writer",
			SkipCheck: true,
		},
		{
			Name: "google_logging_organization_sink.basic",
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

func TestLoggingOrgSink_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_logging_organization_sink" "basic" {
			name   = "basic"
			description = "what it is"
			org_id = "123456789"
		
			destination = "storage.googleapis.com/${google_storage_bucket.log-bucket.name}"
		}
		
		resource "google_storage_bucket" "log-bucket" {
			name = "organization-logging-bucket"
		}
		
		resource "google_project_iam_member" "log-writer" {
			role = "roles/storage.objectCreator"
		
			member = google_logging_organization_sink.basic.writer_identity
		}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"google_logging_organization_sink.basic": map[string]interface{}{
			"monthly_logging_data_gb": 100,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "google_storage_bucket.log-bucket",
			SkipCheck: true,
		},
		{
			Name:      "google_project_iam_member.log-writer",
			SkipCheck: true,
		},
		{
			Name: "google_logging_organization_sink.basic",
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
