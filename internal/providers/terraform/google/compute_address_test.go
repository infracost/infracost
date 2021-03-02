package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestComputeAddress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_compute_address" "static" {
			name = "ipv4-address"
		}
		resource "google_compute_address" "internal" {
			name = "ipv4-address-internal"
			address_type = "INTERNAL"
		}
		resource "google_compute_global_address" "default" {
			name = "global-appserver-ip"
		}
		`
	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_compute_address.static",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "IP address (if used by standard VM)",
					PriceHash:        "63d43e05c6de193d46ac984c5d047c4e-75ba4eb307fdd3d2d30cb3abe7436559",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "IP address (if used by preemptible VM)",
					PriceHash:        "2ec0a063efa9b4e610e5205f9441dc4d-ef2cadbde566a742ff14834f883bcb8a",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "IP address (if unused)",
					PriceHash:        "2aa962ad3e313d7a01f2ea2b98a3cb40-d7883856ef5a8d377f6fc8b3df05ea7e",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "google_compute_global_address.default",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "IP address (if used by standard VM)",
					PriceHash:        "63d43e05c6de193d46ac984c5d047c4e-75ba4eb307fdd3d2d30cb3abe7436559",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "IP address (if used by preemptible VM)",
					PriceHash:        "2ec0a063efa9b4e610e5205f9441dc4d-ef2cadbde566a742ff14834f883bcb8a",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "IP address (if unused)",
					PriceHash:        "2aa962ad3e313d7a01f2ea2b98a3cb40-d7883856ef5a8d377f6fc8b3df05ea7e",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}
	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
