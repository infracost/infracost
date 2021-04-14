package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestNewSQLInstance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "google_sql_database_instance" "sql_server" {
			name             = "master-instance"
			database_version = "SQLSERVER_2017_ENTERPRISE"
			settings {
				tier = "db-custom-16-61440"
				availability_type = "ZONAL"
			}
		}

		resource "google_sql_database_instance" "custom_postgres" {
			name             = "master-instance"
			database_version = "POSTGRES_11"
		
			settings {
				tier = "db-custom-2-13312"
			}
		}

		resource "google_sql_database_instance" "HA_custom_postgres" {
			name             = "master-instance"
			database_version = "POSTGRES_11"
		
			settings {
				tier = "db-custom-16-61440"
			}
		}

		resource "google_sql_database_instance" "HA_small_mysql" {
			name             = "master-instance"
			database_version = "MYSQL_8_0"
		
			settings {
				tier = "db-g1-small"
				availability_type = "REGIONAL"
				disk_size = "100"
			}
		}

		resource "google_sql_database_instance" "small_mysql" {
			name             = "master-instance"
			database_version = "MYSQL_8_0"
		
			settings {
				tier = "db-g1-small"
				availability_type = "ZONAL"
			}
		}

		resource "google_sql_database_instance" "with_replica" {
			name             = "master-instance"
			database_version = "POSTGRES_11"
			settings {
				tier = "db-custom-16-61440"
				availability_type = "REGIONAL"
				disk_size = 500
			}
			replica_configuration {
				username = "replica"
			}
		}
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_sql_database_instance.sql_server",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "vCPUs (zonal)",
					PriceHash:       "b9219a9fb93a15f0d47e45910be1825f-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(16)),
				},
				{
					Name:            "Memory (zonal)",
					PriceHash:       "0d1ee1c196054c47dbd8cc2acbc4d35e-e400b4debea1ba77ad9bec422eeaf576",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(60)),
				},
				{
					Name:            "Storage (SSD, zonal)",
					PriceHash:       "5cf7faa740d422ad2a42937d73517ba4-57bc5d148491a8381abaccb21ca6b4e9",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10)),
				},
			},
		},
		{
			Name: "google_sql_database_instance.custom_postgres",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "vCPUs (zonal)",
					PriceHash:       "8ca4313075dce70bb4b2473b012ddb7c-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(2)),
				},
				{
					Name:            "Memory (zonal)",
					PriceHash:       "d2d3ec68f097b7a7c2302e33ef61b23d-e400b4debea1ba77ad9bec422eeaf576",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(13)),
				},
				{
					Name:            "Storage (SSD, zonal)",
					PriceHash:       "13d1eea521d50ece4069df73fab5e4fe-57bc5d148491a8381abaccb21ca6b4e9",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10)),
				},
				{
					Name:            "SQL instance (db-custom-2-13312, zonal)",
					PriceHash:       "896b2e1bfdcef6f44ba586aea1d0daa1-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "google_sql_database_instance.HA_custom_postgres",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "vCPUs (zonal)",
					PriceHash:       "8ca4313075dce70bb4b2473b012ddb7c-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(16)),
				},
				{
					Name:            "Memory (zonal)",
					PriceHash:       "d2d3ec68f097b7a7c2302e33ef61b23d-e400b4debea1ba77ad9bec422eeaf576",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(60)),
				},
				{
					Name:            "Storage (SSD, zonal)",
					PriceHash:       "13d1eea521d50ece4069df73fab5e4fe-57bc5d148491a8381abaccb21ca6b4e9",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10)),
				},
				{
					Name:            "SQL instance (db-custom-16-61440, zonal)",
					PriceHash:       "350175da373d26dda5a9d4ddba8e37c9-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "google_sql_database_instance.HA_small_mysql",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Storage (SSD, regional)",
					PriceHash:       "ac54437f3c0f733bc629b6cf667e2943-57bc5d148491a8381abaccb21ca6b4e9",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100)),
				},
				{
					Name:            "SQL instance (db-g1-small, regional)",
					PriceHash:       "cc7a2cc3b784c524185839fbbc426b4f-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "google_sql_database_instance.small_mysql",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Storage (SSD, zonal)",
					PriceHash:       "e02fe49c6a08383eadddbc68669618d5-57bc5d148491a8381abaccb21ca6b4e9",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10)),
				},
				{
					Name:            "SQL instance (db-g1-small, zonal)",
					PriceHash:       "107408a8d6858ac0ca4cecf164263aee-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "google_sql_database_instance.with_replica",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "vCPUs (regional)",
					PriceHash:       "06aab83b7c71ff48fedf5b3e3b642901-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(16)),
				},
				{
					Name:            "Memory (regional)",
					PriceHash:       "41ffc843143c058ef633c8f353d88b7e-e400b4debea1ba77ad9bec422eeaf576",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(60)),
				},
				{
					Name:            "Storage (SSD, regional)",
					PriceHash:       "cb35b710d2bdb84771a140c1722c2bc2-57bc5d148491a8381abaccb21ca6b4e9",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(500)),
				},
				{
					Name:            "SQL instance (db-custom-16-61440, regional)",
					PriceHash:       "a1d8913462087d200ce33ea36a8ae5e1-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "Replica",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "vCPUs (zonal)",
							PriceHash:       "8ca4313075dce70bb4b2473b012ddb7c-ef2cadbde566a742ff14834f883bcb8a",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(16)),
						},
						{
							Name:            "Memory (zonal)",
							PriceHash:       "d2d3ec68f097b7a7c2302e33ef61b23d-e400b4debea1ba77ad9bec422eeaf576",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(60)),
						},
						{
							Name:            "Storage (SSD, zonal)",
							PriceHash:       "13d1eea521d50ece4069df73fab5e4fe-57bc5d148491a8381abaccb21ca6b4e9",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(500)),
						},
						{
							Name:            "SQL instance (db-custom-16-61440, zonal)",
							PriceHash:       "350175da373d26dda5a9d4ddba8e37c9-ef2cadbde566a742ff14834f883bcb8a",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
						},
					},
				},
			},
		},
	}
	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestNewSQLInstance_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "google_sql_database_instance" "small_mysql" {
		name             = "master-instance"
		database_version = "MYSQL_8_0"
	
		settings {
			tier = "db-g1-small"
			availability_type = "ZONAL"
		}
	}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"google_sql_database_instance.small_mysql": map[string]interface{}{
			"backup_storage_gb": 100,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_sql_database_instance.small_mysql",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Storage (SSD, zonal)",
					PriceHash:       "e02fe49c6a08383eadddbc68669618d5-57bc5d148491a8381abaccb21ca6b4e9",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10)),
				},
				{
					Name:            "SQL instance (db-g1-small, zonal)",
					PriceHash:       "107408a8d6858ac0ca4cecf164263aee-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "Backups",
					PriceHash:       "5d5ace3b30ea029049048c6fba8d6ce2-57bc5d148491a8381abaccb21ca6b4e9",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
