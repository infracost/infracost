package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
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
		`
	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_compute_address.static",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Static and ephemeral IP addresses in use on standard VM instances",
					PriceHash:        "63d43e05c6de193d46ac984c5d047c4e-92a41b8ee8a64d671e700c781c365c10",
					MonthlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}
	tftest.ResourceTests(t, tf, resourceChecks)
}
