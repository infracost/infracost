package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestComputeRouterNAT(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_compute_router_nat" "nat" {
			name   = "example"
			router = "example-router"
			region = "us-central1"
			nat_ip_allocate_option = "MANUAL_ONLY"
			source_subnetwork_ip_ranges_to_nat = "LIST_OF_SUBNETWORKS"
		}

		resource "google_compute_router_nat" "over_32_vms" {
			name   = "example-over-32-vms"
			router = "example-router"
			region = "us-central1"
			nat_ip_allocate_option = "MANUAL_ONLY"
			source_subnetwork_ip_ranges_to_nat = "LIST_OF_SUBNETWORKS"
		}
	`

	usage := schema.NewUsageMap(map[string]interface{}{
		"google_compute_router_nat.nat": map[string]interface{}{
			"assigned_vms":              4,
			"monthly_data_processed_gb": 1000,
		},
		"google_compute_router_nat.over_32_vms": map[string]interface{}{
			"assigned_vms":              32,
			"monthly_data_processed_gb": 1000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_compute_router_nat.nat",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Assigned VMs (first 32)",
					PriceHash:       "0d11cd10fb408a6cd1c7978fd45dfe0f-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(4)),
				},
				{
					Name:             "Data processed",
					PriceHash:        "b9cdd27fb02db0c665ae0620ebf299ac-8012a4febcd0213911ed09e53341a976",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1000)),
				},
			},
		},
		{
			Name: "google_compute_router_nat.over_32_vms",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Assigned VMs (first 32)",
					PriceHash:       "0d11cd10fb408a6cd1c7978fd45dfe0f-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(32)),
				},
				{
					Name:             "Data processed",
					PriceHash:        "b9cdd27fb02db0c665ae0620ebf299ac-8012a4febcd0213911ed09e53341a976",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1000)),
				},
			},
		},
	}
	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
