package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestContainerNodePool_zonal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_container_cluster" "default" {
			name     = "default"
			location = "us-central1-a"
		}

		resource "google_container_cluster" "node_locations" {
			name           = "node-locations"
			location       = "us-central1-a"

			node_locations = [
				"us-central1-a",
				"us-central1-b",
			]
		}

		resource "google_container_node_pool" "default" {
			name     = "default"
			cluster  = google_container_cluster.default.id
		}

		resource "google_container_node_pool" "with_node_config" {
			name       = "with-node-config"
			cluster    = google_container_cluster.default.id
			node_count = 3

			node_config {
				machine_type    = "n1-standard-16"
				disk_size_gb    = 120
				disk_type       = "pd-ssd"
				local_ssd_count = 1

				guest_accelerator {
					type = "nvidia-tesla-k80"
					count = 4
				}
			}
		}

		resource "google_container_node_pool" "cluster_node_locations" {
			name       = "cluster-node-locations"
			cluster    = google_container_cluster.node_locations.id
			node_count = 2
		}

		resource "google_container_node_pool" "node_locations" {
			name       = "node-locations"
			cluster    = google_container_cluster.default.id
			node_count = 2

			node_locations = [
				"us-central1-a",
				"us-central1-b",
			]
		}

		resource "google_container_node_pool" "initial_node_count" {
			name               = "initial-node-count"
			cluster            = google_container_cluster.default.id
			initial_node_count = 4
		}

		resource "google_container_node_pool" "autoscaling" {
			name    = "autoscaling"
			cluster = google_container_cluster.default.id

			autoscaling {
				min_node_count = 2
				max_node_count = 10
			}
		}
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "google_container_cluster.default",
			SkipCheck: true,
		},
		{
			Name:      "google_container_cluster.node_locations",
			SkipCheck: true,
		},
		{
			Name: "google_container_node_pool.default",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, on-demand, e2-medium)",
					PriceHash:       "1ed2c16a0b9da97ed123b5266cba4e50-66d0d770bee368b4f2a8f2f597eeb417",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 3)),
				},
				{
					Name:             "Standard provisioned storage (pd-standard)",
					PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 3)),
				},
			},
		},
		{
			Name: "google_container_node_pool.with_node_config",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Instance usage (Linux/UNIX, on-demand, n1-standard-16)",
					PriceHash:        "f9362669032dbf3ed07fe0340744d593-66d0d770bee368b4f2a8f2f597eeb417",
					HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 3)),
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730 * 0.7 * 3)),
				},
				{
					Name:             "SSD provisioned storage (pd-ssd)",
					PriceHash:        "7317191236b3f20b4e8122bddb65e5cf-57bc5d148491a8381abaccb21ca6b4e9",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(120 * 3)),
				},
				{
					Name:             "Local SSD provisioned storage",
					PriceHash:        "dae3672d3f7605d4e5c6d48aa342d66c-57bc5d148491a8381abaccb21ca6b4e9",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(375 * 3)),
				},
				{
					Name:             "NVIDIA Tesla K80 (on-demand)",
					PriceHash:        "6ee67450f39db80d8bd818ed6a12da85-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(4 * 3)),
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(4 * 730 * 0.7 * 3)),
				},
			},
		},
		{
			Name: "google_container_node_pool.cluster_node_locations",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, on-demand, e2-medium)",
					PriceHash:       "1ed2c16a0b9da97ed123b5266cba4e50-66d0d770bee368b4f2a8f2f597eeb417",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 4)),
				},
				{
					Name:             "Standard provisioned storage (pd-standard)",
					PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 4)),
				},
			},
		},
		{
			Name: "google_container_node_pool.node_locations",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, on-demand, e2-medium)",
					PriceHash:       "1ed2c16a0b9da97ed123b5266cba4e50-66d0d770bee368b4f2a8f2f597eeb417",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 4)),
				},
				{
					Name:             "Standard provisioned storage (pd-standard)",
					PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 4)),
				},
			},
		},
		{
			Name: "google_container_node_pool.initial_node_count",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, on-demand, e2-medium)",
					PriceHash:       "1ed2c16a0b9da97ed123b5266cba4e50-66d0d770bee368b4f2a8f2f597eeb417",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 4)),
				},
				{
					Name:             "Standard provisioned storage (pd-standard)",
					PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 4)),
				},
			},
		},
		{
			Name: "google_container_node_pool.autoscaling",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, on-demand, e2-medium)",
					PriceHash:       "1ed2c16a0b9da97ed123b5266cba4e50-66d0d770bee368b4f2a8f2f597eeb417",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 2)),
				},
				{
					Name:             "Standard provisioned storage (pd-standard)",
					PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 2)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestContainerNodePool_regional(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_container_cluster" "default" {
			name     = "default"
			location = "us-central1"
		}

		resource "google_container_cluster" "node_locations" {
			name           = "node-locations"
			location       = "us-central1"

			node_locations = [
				"us-central1-a",
				"us-central1-b",
			]
		}

		resource "google_container_node_pool" "default" {
			name     = "default"
			cluster  = google_container_cluster.default.id
		}

		resource "google_container_node_pool" "cluster_node_locations" {
			name       = "cluster-node-locations"
			cluster    = google_container_cluster.node_locations.id
			node_count = 2
		}

		resource "google_container_node_pool" "node_locations" {
			name       = "node-locations"
			cluster    = google_container_cluster.default.id
			node_count = 2

			node_locations = [
				"us-central1-a",
				"us-central1-b",
			]
		}

		resource "google_container_node_pool" "initial_node_count" {
			name               = "initial-node-count"
			cluster            = google_container_cluster.default.id
			initial_node_count = 4
		}

		resource "google_container_node_pool" "autoscaling" {
			name    = "autoscaling"
			cluster = google_container_cluster.default.id

			autoscaling {
				min_node_count = 2
				max_node_count = 10
			}
		}
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "google_container_cluster.default",
			SkipCheck: true,
		},
		{
			Name:      "google_container_cluster.node_locations",
			SkipCheck: true,
		},
		{
			Name: "google_container_node_pool.default",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, on-demand, e2-medium)",
					PriceHash:       "1ed2c16a0b9da97ed123b5266cba4e50-66d0d770bee368b4f2a8f2f597eeb417",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 9)),
				},
				{
					Name:             "Standard provisioned storage (pd-standard)",
					PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 9)),
				},
			},
		},
		{
			Name: "google_container_node_pool.cluster_node_locations",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, on-demand, e2-medium)",
					PriceHash:       "1ed2c16a0b9da97ed123b5266cba4e50-66d0d770bee368b4f2a8f2f597eeb417",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 4)),
				},
				{
					Name:             "Standard provisioned storage (pd-standard)",
					PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 4)),
				},
			},
		},
		{
			Name: "google_container_node_pool.node_locations",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, on-demand, e2-medium)",
					PriceHash:       "1ed2c16a0b9da97ed123b5266cba4e50-66d0d770bee368b4f2a8f2f597eeb417",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 4)),
				},
				{
					Name:             "Standard provisioned storage (pd-standard)",
					PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 4)),
				},
			},
		},
		{
			Name: "google_container_node_pool.initial_node_count",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, on-demand, e2-medium)",
					PriceHash:       "1ed2c16a0b9da97ed123b5266cba4e50-66d0d770bee368b4f2a8f2f597eeb417",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 12)),
				},
				{
					Name:             "Standard provisioned storage (pd-standard)",
					PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 12)),
				},
			},
		},
		{
			Name: "google_container_node_pool.autoscaling",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, on-demand, e2-medium)",
					PriceHash:       "1ed2c16a0b9da97ed123b5266cba4e50-66d0d770bee368b4f2a8f2f597eeb417",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 6)),
				},
				{
					Name:             "Standard provisioned storage (pd-standard)",
					PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 6)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestContainerNodePool_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_container_cluster" "zonal" {
			name     = "default"
			location = "us-central1-a"
		}

		resource "google_container_cluster" "regional" {
			name     = "default"
			location = "us-central1"
		}

		resource "google_container_cluster" "node_locations" {
			name     = "node-locations"
			location = "us-central1"

			node_locations = [
				"us-central1-a",
				"us-central1-b",
			]
		}

		resource "google_container_node_pool" "zonal" {
			name       = "zonal"
			cluster    = google_container_cluster.zonal.id
			node_count = 3
		}

		resource "google_container_node_pool" "regional" {
			name       = "regional"
			cluster    = google_container_cluster.regional.id
			node_count = 3
		}

		resource "google_container_node_pool" "node_locations" {
			name       = "node-locations"
			cluster    = google_container_cluster.node_locations.id
			node_count = 3
		}
	`

	usage := schema.NewUsageMap(map[string]interface{}{
		"google_container_node_pool.zonal": map[string]interface{}{
			"nodes": 4,
		},
		"google_container_node_pool.regional": map[string]interface{}{
			"nodes": 4,
		},
		"google_container_node_pool.node_locations": map[string]interface{}{
			"nodes": 4,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "google_container_cluster.zonal",
			SkipCheck: true,
		},
		{
			Name:      "google_container_cluster.regional",
			SkipCheck: true,
		},
		{
			Name:      "google_container_cluster.node_locations",
			SkipCheck: true,
		},
		{
			Name: "google_container_node_pool.zonal",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, on-demand, e2-medium)",
					PriceHash:       "1ed2c16a0b9da97ed123b5266cba4e50-66d0d770bee368b4f2a8f2f597eeb417",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 4)),
				},
				{
					Name:             "Standard provisioned storage (pd-standard)",
					PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 4)),
				},
			},
		},
		{
			Name: "google_container_node_pool.regional",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, on-demand, e2-medium)",
					PriceHash:       "1ed2c16a0b9da97ed123b5266cba4e50-66d0d770bee368b4f2a8f2f597eeb417",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 12)),
				},
				{
					Name:             "Standard provisioned storage (pd-standard)",
					PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 12)),
				},
			},
		},
		{
			Name: "google_container_node_pool.node_locations",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (Linux/UNIX, on-demand, e2-medium)",
					PriceHash:       "1ed2c16a0b9da97ed123b5266cba4e50-66d0d770bee368b4f2a8f2f597eeb417",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 8)),
				},
				{
					Name:             "Standard provisioned storage (pd-standard)",
					PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 8)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
