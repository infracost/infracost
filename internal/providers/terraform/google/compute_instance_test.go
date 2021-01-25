package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestComputeInstance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_compute_instance" "standard" {
			name         = "standard"
			machine_type = "f1-micro"
			zone         = "us-central1-a"

			boot_disk {
				initialize_params {
					image = "centos-cloud/centos-7"
				}
			}

			network_interface {
				network = "default"
			}
		}

		resource "google_compute_instance" "ssd" {
			name         = "ssd"
			machine_type = "f1-micro"
			zone         = "us-central1-a"

			boot_disk {
				initialize_params {
					image = "centos-cloud/centos-7"
					size = 40
					type = "pd-ssd"
				}
			}

			network_interface {
				network = "default"
			}
		}

		resource "google_compute_instance" "preemptible" {
			name         = "preemptible"
			machine_type = "f1-micro"
			zone         = "us-central1-a"

			boot_disk {
				initialize_params {
					image = "centos-cloud/centos-7"
				}
			}

			scheduling {
				preemptible = true
			}

			network_interface {
				network = "default"
			}
		}

		resource "google_compute_instance" "local_ssd" {
			name         = "local_ssd"
			machine_type = "f1-micro"
			zone         = "us-central1-a"

			boot_disk {
				initialize_params {
					image = "centos-cloud/centos-7"
				}
			}

			network_interface {
				network = "default"
			}

			scratch_disk {
				interface = "SCSI"
			}

			scratch_disk {
				interface = "SCSI"
			}
		}

		resource "google_compute_instance" "preemptible_local_ssd" {
			name         = "preemptible_local_ssd"
			machine_type = "f1-micro"
			zone         = "us-central1-a"

			boot_disk {
				initialize_params {
					image = "centos-cloud/centos-7"
				}
			}

			network_interface {
				network = "default"
			}

			scheduling {
				preemptible = true
			}

			scratch_disk {
				interface = "SCSI"
			}
		}


		resource "google_compute_instance" "gpu" {
			name         = "gpu"
			machine_type = "n1-standard-16"
			zone         = "us-central1-a"

			boot_disk {
				initialize_params {
					image = "centos-cloud/centos-7"
				}
			}

			guest_accelerator {
				type = "nvidia-tesla-k80"
				count = 4
			}

			network_interface {
				network = "default"
			}
		}

		resource "google_compute_instance" "preemptible_gpu" {
			name         = "preemptible_gpu"
			machine_type = "n1-standard-16"
			zone         = "us-central1-a"

			boot_disk {
				initialize_params {
					image = "centos-cloud/centos-7"
				}
			}

			guest_accelerator {
				type = "nvidia-tesla-k80"
				count = 4
			}

			scheduling {
				preemptible = true
			}

			network_interface {
				network = "default"
			}
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_compute_instance.standard",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Linux/UNIX usage (on-demand, f1-micro)",
					PriceHash:        "7b4212f1f3122457b7bc03baa4c3acaf-66d0d770bee368b4f2a8f2f597eeb417",
					HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730 * 0.7)),
				},
				{
					Name:             "Standard provisioned storage (pd-standard)",
					PriceHash:        "4e58b7b536714dfce35b3050caa6034b-af6a951f170fc579633ad2c8f86a9dca",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10)),
				},
			},
		},
		{
			Name: "google_compute_instance.ssd",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "Linux/UNIX usage (on-demand, f1-micro)",
					SkipCheck: true,
				},
				{
					Name:             "SSD provisioned storage (pd-ssd)",
					PriceHash:        "7317191236b3f20b4e8122bddb65e5cf-57bc5d148491a8381abaccb21ca6b4e9",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(40)),
				},
			},
		},
		{
			Name: "google_compute_instance.preemptible",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Linux/UNIX usage (preemptible, f1-micro)",
					PriceHash:        "7b4212f1f3122457b7bc03baa4c3acaf-cfd7416b9a6fd4bc337fd81f1974337e",
					HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(730)),
				},
				{
					Name:      "Standard provisioned storage (pd-standard)",
					SkipCheck: true,
				},
			},
		},
		{
			Name: "google_compute_instance.local_ssd",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "Linux/UNIX usage (on-demand, f1-micro)",
					SkipCheck: true,
				},
				{
					Name:      "Standard provisioned storage (pd-standard)",
					SkipCheck: true,
				},
				{
					Name:             "Local SSD provisioned storage",
					PriceHash:        "dae3672d3f7605d4e5c6d48aa342d66c-57bc5d148491a8381abaccb21ca6b4e9",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(750)),
				},
			},
		},
		{
			Name: "google_compute_instance.preemptible_local_ssd",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "Linux/UNIX usage (preemptible, f1-micro)",
					SkipCheck: true,
				},
				{
					Name:      "Standard provisioned storage (pd-standard)",
					SkipCheck: true,
				},
				{
					Name:             "Local SSD provisioned storage",
					PriceHash:        "98ab5dbbbbe93cabb1e1e1d980508030-eeeaf87a70d307299f72242c0b164882",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(375)),
				},
			},
		},
		{
			Name: "google_compute_instance.gpu",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "Linux/UNIX usage (on-demand, n1-standard-16)",
					SkipCheck: true,
				},
				{
					Name:      "Standard provisioned storage (pd-standard)",
					SkipCheck: true,
				},
				{
					Name:             "NVIDIA Tesla K80 (on-demand)",
					PriceHash:        "6ee67450f39db80d8bd818ed6a12da85-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(4)),
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(4 * 730 * 0.7)),
				},
			},
		},
		{
			Name: "google_compute_instance.preemptible_gpu",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "Linux/UNIX usage (preemptible, n1-standard-16)",
					SkipCheck: true,
				},
				{
					Name:      "Standard provisioned storage (pd-standard)",
					SkipCheck: true,
				},
				{
					Name:             "NVIDIA Tesla K80 (preemptible)",
					PriceHash:        "e5560cc54e4a7bae9be88923f2f37fd1-6f1660572afcedbb83d7821be40f13a3",
					HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(4)),
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(4 * 730)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
