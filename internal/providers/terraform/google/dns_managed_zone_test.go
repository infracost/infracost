package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestDNSManagedZone(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_dns_managed_zone" "zone" {
			name        = "example"
			dns_name    = "example-123.com."
		}
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_dns_managed_zone.zone",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Managed zone",
					PriceHash:        "ab8fb71420449113cb43ba3a1a381b24-beb6f6e797864c7d03c4cb4301c625b4",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}
	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
