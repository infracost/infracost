package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestComputeMachineImage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "google_compute_instance" "vm" {
		name         = "vm"
		machine_type = "e2-medium"
	
		boot_disk {
			initialize_params {
				image = "fake"
			}
		}
	
		network_interface {
			network = "fake"
		}
	}
	
	resource "google_compute_machine_image" "image" {
		provider     		= "google-beta"
		name            = "image"
		source_instance = google_compute_instance.vm.self_link
	}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "google_compute_instance.vm",
			SkipCheck: true,
		},
		{
			Name: "google_compute_machine_image.image",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage",
					PriceHash:        "26be44612e84631670a69db241945bfb-57bc5d148491a8381abaccb21ca6b4e9",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestComputeMachineImage_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "google_compute_instance" "vm" {
		name         = "vm"
		machine_type = "e2-medium"
		zone         = "us-central1-a"
	
		boot_disk {
			initialize_params {
				image = "fake"
			}
		}
	
		network_interface {
			network = "fake"
		}
	}
	
	resource "google_compute_machine_image" "image" {
		provider 				= "google-beta"
		name            = "image"
		source_instance = google_compute_instance.vm.self_link
	}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"google_compute_machine_image.image": map[string]interface{}{
			"storage_gb": 100,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "google_compute_instance.vm",
			SkipCheck: true,
		},
		{
			Name: "google_compute_machine_image.image",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage",
					PriceHash:        "26be44612e84631670a69db241945bfb-57bc5d148491a8381abaccb21ca6b4e9",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
