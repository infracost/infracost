package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestContainerCluster(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_container_cluster" "zonal" {
			name     = "zonal"
			location = "us-central1-a"
		}

		resource "google_container_cluster" "regional" {
			name     = "regional"
			location = "us-central1"
		}

		resource "google_container_cluster" "node_locations" {
			name     = "node-locations"
			location = "us-central1"

			node_locations = [
				"us-central1-a",
				"us-central1-b"
			]
		}

		resource "google_container_cluster" "with_node_config" {
			name               = "with-node-config"
			location           = "us-central1-a"
			initial_node_count = 3

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

		resource "google_container_cluster" "with_node_pools_zonal" {
			name     = "with-node-pools"
			location = "us-central1-a"

			node_pool {
				node_count = 2

				node_config {
					machine_type = "n1-standard-16"
				}
			}

			node_pool {
				node_count = 4

				node_config {
					machine_type  = "n1-standard-16"
					preemptible   = true
				}
			}
		}

		resource "google_container_cluster" "with_node_pools_regional" {
			name     = "with-node-pools-regional"
			location = "us-central1"

			node_pool {
				node_count = 2

				node_config {
					machine_type = "n1-standard-16"
				}
			}

			node_pool {
				node_count = 4

				node_config {
					machine_type  = "n1-standard-16"
					preemptible   = true
				}
			}
		}

		resource "google_container_cluster" "with_node_pools_node_locations" {
			name     = "with-node-pools-regional"
			location = "us-central1"

			node_locations = [
				"us-central1-a",
				"us-central1-b"
			]

			node_pool {
				node_count = 2

				node_config {
					machine_type = "n1-standard-16"
				}
			}

			node_pool {
				node_count = 2

				node_locations = [
					"us-central1-a"
				]

				node_config {
					machine_type  = "n1-standard-16"
					preemptible   = true
				}
			}
		}
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_container_cluster.zonal",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Cluster management fee",
					PriceHash:       "8f899c90440972565d0f2d5b8ff11ae0-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "default_pool",
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
			},
		},
		{
			Name: "google_container_cluster.regional",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Cluster management fee",
					PriceHash:       "141881f456123b9e5e0538f11e995678-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "default_pool",
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
			},
		},
		{
			Name: "google_container_cluster.node_locations",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Cluster management fee",
					PriceHash:       "141881f456123b9e5e0538f11e995678-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "default_pool",
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
			},
		},
		{
			Name: "google_container_cluster.with_node_config",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Cluster management fee",
					PriceHash:       "8f899c90440972565d0f2d5b8ff11ae0-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "default_pool",
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
			},
		},
		{
			Name: "google_container_cluster.with_node_pools_zonal",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Cluster management fee",
					PriceHash:       "8f899c90440972565d0f2d5b8ff11ae0-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "default_pool",
					SkipCheck: true,
				},
				{
					Name: "node_pool[0]",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:             "Instance usage (Linux/UNIX, on-demand, n1-standard-16)",
							PriceHash:        "f9362669032dbf3ed07fe0340744d593-66d0d770bee368b4f2a8f2f597eeb417",
							HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 2)),
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730 * 0.7 * 2)),
						},
						{
							Name:             "Standard provisioned storage (pd-standard)",
							PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 2)),
						},
					},
				},
				{
					Name: "node_pool[1]",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:             "Instance usage (Linux/UNIX, preemptible, n1-standard-16)",
							PriceHash:        "f9362669032dbf3ed07fe0340744d593-cfd7416b9a6fd4bc337fd81f1974337e",
							HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 4)),
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730 * 4)),
						},
						{
							Name:             "Standard provisioned storage (pd-standard)",
							PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 4)),
						},
					},
				},
			},
		},
		{
			Name: "google_container_cluster.with_node_pools_regional",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Cluster management fee",
					PriceHash:       "141881f456123b9e5e0538f11e995678-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "default_pool",
					SkipCheck: true,
				},
				{
					Name: "node_pool[0]",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:             "Instance usage (Linux/UNIX, on-demand, n1-standard-16)",
							PriceHash:        "f9362669032dbf3ed07fe0340744d593-66d0d770bee368b4f2a8f2f597eeb417",
							HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 6)),
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730 * 0.7 * 6)),
						},
						{
							Name:             "Standard provisioned storage (pd-standard)",
							PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 6)),
						},
					},
				},
				{
					Name: "node_pool[1]",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:             "Instance usage (Linux/UNIX, preemptible, n1-standard-16)",
							PriceHash:        "f9362669032dbf3ed07fe0340744d593-cfd7416b9a6fd4bc337fd81f1974337e",
							HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 12)),
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730 * 12)),
						},
						{
							Name:             "Standard provisioned storage (pd-standard)",
							PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 12)),
						},
					},
				},
			},
		},
		{
			Name: "google_container_cluster.with_node_pools_node_locations",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Cluster management fee",
					PriceHash:       "141881f456123b9e5e0538f11e995678-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "default_pool",
					SkipCheck: true,
				},
				{
					Name: "node_pool[0]",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:             "Instance usage (Linux/UNIX, on-demand, n1-standard-16)",
							PriceHash:        "f9362669032dbf3ed07fe0340744d593-66d0d770bee368b4f2a8f2f597eeb417",
							HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 4)),
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730 * 0.7 * 4)),
						},
						{
							Name:             "Standard provisioned storage (pd-standard)",
							PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 4)),
						},
					},
				},
				{
					Name: "node_pool[1]",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:             "Instance usage (Linux/UNIX, preemptible, n1-standard-16)",
							PriceHash:        "f9362669032dbf3ed07fe0340744d593-cfd7416b9a6fd4bc337fd81f1974337e",
							HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 2)),
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730 * 2)),
						},
						{
							Name:             "Standard provisioned storage (pd-standard)",
							PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 2)),
						},
					},
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestContainerCluster_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_container_cluster" "zonal" {
			name               = "zonal"
			location           = "us-central1-a"
			initial_node_count = 3
		}

		resource "google_container_cluster" "regional" {
			name               = "regional"
			location           = "us-central1"
			initial_node_count = 3
		}

		resource "google_container_cluster" "node_locations" {
			name               = "node-locations"
			location           = "us-central1"
			initial_node_count = 3

			node_locations = [
				"us-central1-a",
				"us-central1-b"
			]
		}

		resource "google_container_cluster" "with_node_pools_zonal" {
			name     = "with-node-pools"
			location = "us-central1-a"

			node_pool {
				node_count = 2

				node_config {
					machine_type = "n1-standard-16"
				}
			}

			node_pool {
				node_count = 4

				node_config {
					machine_type  = "n1-standard-16"
					preemptible   = true
				}
			}
		}

		resource "google_container_cluster" "with_node_pools_regional" {
			name     = "with-node-pools-regional"
			location = "us-central1"

			node_pool {
				node_count = 2

				node_config {
					machine_type = "n1-standard-16"
				}
			}

			node_pool {
				node_count = 4

				node_config {
					machine_type  = "n1-standard-16"
					preemptible   = true
				}
			}
		}

		resource "google_container_cluster" "with_node_pools_node_locations" {
			name     = "with-node-pools-regional"
			location = "us-central1"

			node_locations = [
				"us-central1-a",
				"us-central1-b"
			]

			node_pool {
				node_count = 2

				node_config {
					machine_type = "n1-standard-16"
				}
			}

			node_pool {
				node_count = 2

				node_locations = [
					"us-central1-a"
				]

				node_config {
					machine_type  = "n1-standard-16"
					preemptible   = true
				}
			}
		}
	`

	usage := schema.NewUsageMap(map[string]interface{}{
		"google_container_cluster.zonal": map[string]interface{}{
			"nodes": 4,
		},
		"google_container_cluster.regional": map[string]interface{}{
			"nodes": 4,
		},
		"google_container_cluster.node_locations": map[string]interface{}{
			"nodes": 4,
		},
		"google_container_cluster.with_node_pools_zonal": map[string]interface{}{
			"node_pool[0].nodes": 4,
			"node_pool[1].nodes": 4,
		},
		"google_container_cluster.with_node_pools_regional": map[string]interface{}{
			"node_pool[0].nodes": 4,
			"node_pool[1].nodes": 4,
		},
		"google_container_cluster.with_node_pools_node_locations": map[string]interface{}{
			"node_pool[0].nodes": 4,
			"node_pool[1].nodes": 4,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_container_cluster.zonal",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Cluster management fee",
					PriceHash:       "8f899c90440972565d0f2d5b8ff11ae0-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "default_pool",
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
			},
		},
		{
			Name: "google_container_cluster.regional",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Cluster management fee",
					PriceHash:       "141881f456123b9e5e0538f11e995678-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "default_pool",
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
			},
		},
		{
			Name: "google_container_cluster.node_locations",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Cluster management fee",
					PriceHash:       "141881f456123b9e5e0538f11e995678-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "default_pool",
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
			},
		},
		{
			Name: "google_container_cluster.with_node_pools_zonal",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Cluster management fee",
					PriceHash:       "8f899c90440972565d0f2d5b8ff11ae0-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "default_pool",
					SkipCheck: true,
				},
				{
					Name: "node_pool[0]",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:             "Instance usage (Linux/UNIX, on-demand, n1-standard-16)",
							PriceHash:        "f9362669032dbf3ed07fe0340744d593-66d0d770bee368b4f2a8f2f597eeb417",
							HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 4)),
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730 * 0.7 * 4)),
						},
						{
							Name:             "Standard provisioned storage (pd-standard)",
							PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 4)),
						},
					},
				},
				{
					Name: "node_pool[1]",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:             "Instance usage (Linux/UNIX, preemptible, n1-standard-16)",
							PriceHash:        "f9362669032dbf3ed07fe0340744d593-cfd7416b9a6fd4bc337fd81f1974337e",
							HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 4)),
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730 * 4)),
						},
						{
							Name:             "Standard provisioned storage (pd-standard)",
							PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 4)),
						},
					},
				},
			},
		},
		{
			Name: "google_container_cluster.with_node_pools_regional",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Cluster management fee",
					PriceHash:       "141881f456123b9e5e0538f11e995678-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "default_pool",
					SkipCheck: true,
				},
				{
					Name: "node_pool[0]",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:             "Instance usage (Linux/UNIX, on-demand, n1-standard-16)",
							PriceHash:        "f9362669032dbf3ed07fe0340744d593-66d0d770bee368b4f2a8f2f597eeb417",
							HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 12)),
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730 * 0.7 * 12)),
						},
						{
							Name:             "Standard provisioned storage (pd-standard)",
							PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 12)),
						},
					},
				},
				{
					Name: "node_pool[1]",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:             "Instance usage (Linux/UNIX, preemptible, n1-standard-16)",
							PriceHash:        "f9362669032dbf3ed07fe0340744d593-cfd7416b9a6fd4bc337fd81f1974337e",
							HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 12)),
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730 * 12)),
						},
						{
							Name:             "Standard provisioned storage (pd-standard)",
							PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 12)),
						},
					},
				},
			},
		},
		{
			Name: "google_container_cluster.with_node_pools_node_locations",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Cluster management fee",
					PriceHash:       "141881f456123b9e5e0538f11e995678-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "default_pool",
					SkipCheck: true,
				},
				{
					Name: "node_pool[0]",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:             "Instance usage (Linux/UNIX, on-demand, n1-standard-16)",
							PriceHash:        "f9362669032dbf3ed07fe0340744d593-66d0d770bee368b4f2a8f2f597eeb417",
							HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 8)),
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730 * 0.7 * 8)),
						},
						{
							Name:             "Standard provisioned storage (pd-standard)",
							PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 8)),
						},
					},
				},
				{
					Name: "node_pool[1]",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:             "Instance usage (Linux/UNIX, preemptible, n1-standard-16)",
							PriceHash:        "f9362669032dbf3ed07fe0340744d593-cfd7416b9a6fd4bc337fd81f1974337e",
							HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1 * 4)),
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730 * 4)),
						},
						{
							Name:             "Standard provisioned storage (pd-standard)",
							PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100 * 4)),
						},
					},
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
