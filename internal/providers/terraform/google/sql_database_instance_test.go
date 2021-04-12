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
	  resource "google_sql_database" "database" {
			name     = "my-database"
			instance = google_sql_database_instance.my_sql_instance.name
			}
			
			resource "google_sql_database_instance" "my_sql_instance" {
			name   = "my-database-instance"
			database_version = "SQLSERVER_2017_ENTERPRISE"
			settings {
				tier = "db-f1-micro"
				availability_type = "ZONAL"
				disk_size = 50
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
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_sql_database_instance.my_sql_instance",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "SQL instance (SQL Server, db-f1-micro)",
					PriceHash:       "8c6410e6b05f87ffc7ee2268f2d7afc7-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "Storage (SSD)",
					PriceHash:       "5cf7faa740d422ad2a42937d73517ba4-57bc5d148491a8381abaccb21ca6b4e9",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(50)),
				},
				{
					Name:            "License (Enterprise)",
					PriceHash:       "577c5d61a66cb3b7eaf6b405c1d5f785-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "google_sql_database_instance.custom_postgres",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "CPU Credits",
					PriceHash:       "8ca4313075dce70bb4b2473b012ddb7c-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(2)),
				},
				{
					Name:            "Memory",
					PriceHash:       "d2d3ec68f097b7a7c2302e33ef61b23d-e400b4debea1ba77ad9bec422eeaf576",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(13)),
				},
				{
					Name:            "Storage (SSD)",
					PriceHash:       "13d1eea521d50ece4069df73fab5e4fe-57bc5d148491a8381abaccb21ca6b4e9",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10)),
				},
				{
					Name:            "SQL instance (PostgreSQL, custom)",
					PriceHash:       "896b2e1bfdcef6f44ba586aea1d0daa1-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "google_sql_database_instance.HA_custom_postgres",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "CPU Credits",
					PriceHash:       "8ca4313075dce70bb4b2473b012ddb7c-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(16)),
				},
				{
					Name:            "Memory",
					PriceHash:       "d2d3ec68f097b7a7c2302e33ef61b23d-e400b4debea1ba77ad9bec422eeaf576",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(60)),
				},
				{
					Name:            "Storage (SSD)",
					PriceHash:       "13d1eea521d50ece4069df73fab5e4fe-57bc5d148491a8381abaccb21ca6b4e9",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10)),
				},
				{
					Name:            "SQL instance (PostgreSQL, custom)",
					PriceHash:       "350175da373d26dda5a9d4ddba8e37c9-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "google_sql_database_instance.HA_small_mysql",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Storage (SSD)",
					PriceHash:       "ac54437f3c0f733bc629b6cf667e2943-57bc5d148491a8381abaccb21ca6b4e9",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100)),
				},
				{
					Name:            "SQL instance (MySQL, db-g1-small)",
					PriceHash:       "cc7a2cc3b784c524185839fbbc426b4f-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "google_sql_database_instance.small_mysql",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Storage (SSD)",
					PriceHash:       "e02fe49c6a08383eadddbc68669618d5-57bc5d148491a8381abaccb21ca6b4e9",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10)),
				},
				{
					Name:            "SQL instance (MySQL, db-g1-small)",
					PriceHash:       "107408a8d6858ac0ca4cecf164263aee-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
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
			"monthly_backup_gb": 100,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_sql_database_instance.small_mysql",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Storage (SSD)",
					PriceHash:       "e02fe49c6a08383eadddbc68669618d5-57bc5d148491a8381abaccb21ca6b4e9",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10)),
				},
				{
					Name:            "SQL instance (MySQL, db-g1-small)",
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
