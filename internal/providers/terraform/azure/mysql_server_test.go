package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestMySQLServer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "azurerm_resource_group" "example" {
		name = "example-resources"
		location = "eastus"
	}

	resource "azurerm_mysql_server" "basic_2core" {
		name = "example-mariadb-server"
		location = azurerm_resource_group.example.location
		resource_group_name = azurerm_resource_group.example.name

		administrator_login = "fake"
		administrator_login_password = "fake"

		sku_name = "B_Gen5_2"
		storage_mb = 5120
		version = "5.7"

		geo_redundant_backup_enabled = false
		ssl_enforcement_enabled = true
	}

	resource "azurerm_mysql_server" "gp_4core" {
		name = "example-mariadb-server"
		location = azurerm_resource_group.example.location
		resource_group_name = azurerm_resource_group.example.name

		administrator_login = "fake"
		administrator_login_password = "fake"

		sku_name = "GP_Gen5_4"
		storage_mb = 4096000
		version = "5.7"

		geo_redundant_backup_enabled = false
		ssl_enforcement_enabled = true
	}

	resource "azurerm_mysql_server" "mo_16core" {
		name = "example-mariadb-server"
		location = azurerm_resource_group.example.location
		resource_group_name = azurerm_resource_group.example.name

		administrator_login = "fake"
		administrator_login_password = "fake"

		sku_name = "MO_Gen5_16"
		storage_mb = 5120
		version = "5.7"

		geo_redundant_backup_enabled = false
		ssl_enforcement_enabled = true
	}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "azurerm_resource_group.example",
			SkipCheck: true,
		},
		{
			Name: "azurerm_mysql_server.basic_2core",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (B_Gen5_2)",
					PriceHash:       "25f110f69e5c58202a61fe0ad60dabd0-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "44f2399878d604f3dbe60f8eb46214ef-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(5)),
				},
				{
					Name:             "Additional backup storage",
					PriceHash:        "f1f447e5d09aaf90cffc2c7806a7d09a-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
		{
			Name: "azurerm_mysql_server.gp_4core",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (GP_Gen5_4)",
					PriceHash:       "3c88d2960428405e4b81d7a276fb5d61-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "85e7a7e5af2de2ac5377a9b9e0d37066-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(4000)),
				},
				{
					Name:             "Additional backup storage",
					PriceHash:        "f1f447e5d09aaf90cffc2c7806a7d09a-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
		{
			Name: "azurerm_mysql_server.mo_16core",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (MO_Gen5_16)",
					PriceHash:       "f5f08031d7660eca839675a2c783d0ee-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "85e7a7e5af2de2ac5377a9b9e0d37066-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(5)),
				},
				{
					Name:             "Additional backup storage",
					PriceHash:        "f1f447e5d09aaf90cffc2c7806a7d09a-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestMySQLServer_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "azurerm_resource_group" "example" {
		name = "example-resources"
		location = "eastus"
	}
	resource "azurerm_mysql_server" "without_geo" {
		name = "example-mariadb-server"
		location = azurerm_resource_group.example.location
		resource_group_name = azurerm_resource_group.example.name

		administrator_login = "fake"
		administrator_login_password = "fake"

		sku_name = "MO_Gen5_16"
		storage_mb = 5120
		version = "5.7"

		geo_redundant_backup_enabled = false
		ssl_enforcement_enabled = true
	}

	resource "azurerm_mysql_server" "with_geo" {
		name = "example-mariadb-server"
		location = azurerm_resource_group.example.location
		resource_group_name = azurerm_resource_group.example.name

		administrator_login = "fake"
		administrator_login_password = "fake"

		sku_name = "GP_Gen5_4"
		storage_mb = 4096000
		version = "5.7"

		geo_redundant_backup_enabled = true
		ssl_enforcement_enabled = true
	}
	`

	usage := schema.NewUsageMap(map[string]interface{}{
		"azurerm_mysql_server.without_geo": map[string]interface{}{
			"additional_backup_storage_gb": 2000,
		},
		"azurerm_mysql_server.with_geo": map[string]interface{}{
			"additional_backup_storage_gb": 3000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "azurerm_resource_group.example",
			SkipCheck: true,
		},
		{
			Name: "azurerm_mysql_server.without_geo",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (MO_Gen5_16)",
					PriceHash:       "f5f08031d7660eca839675a2c783d0ee-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "85e7a7e5af2de2ac5377a9b9e0d37066-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(5)),
				},
				{
					Name:             "Additional backup storage",
					PriceHash:        "f1f447e5d09aaf90cffc2c7806a7d09a-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(2000)),
				},
			},
		},
		{
			Name: "azurerm_mysql_server.with_geo",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (GP_Gen5_4)",
					PriceHash:       "3c88d2960428405e4b81d7a276fb5d61-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "85e7a7e5af2de2ac5377a9b9e0d37066-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(4000)),
				},
				{
					Name:             "Additional backup storage",
					PriceHash:        "2e931e393feeb0f33c74d4f35b16c0db-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(3000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
