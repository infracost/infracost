package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestDNSRecordSet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_dns_record_set" "frontend" {
			name = "frontend.123"
			type = "A"
			ttl  = 300
			rrdatas = ["123.123.123.123]"]
			managed_zone = "zone"
		}		
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_dns_record_set.frontend",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Queries",
					PriceHash:        "eceb1d425e4ccbafe3a6d1bdc3b7c93a-d1b3ead1ad60245fc548d1764c127c05",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}
	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestDNSRecordSet_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_dns_record_set" "frontend" {
			name = "frontend.123"
			type = "A"
			ttl  = 300
			rrdatas = ["123.123.123.123]"]
			managed_zone = "zone"
		}		
	`

	usage := schema.NewUsageMap(map[string]interface{}{
		"google_dns_record_set.frontend": map[string]interface{}{
			"monthly_queries": 100000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_dns_record_set.frontend",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Queries",
					PriceHash:        "eceb1d425e4ccbafe3a6d1bdc3b7c93a-d1b3ead1ad60245fc548d1764c127c05",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100000)),
				},
			},
		},
	}
	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
