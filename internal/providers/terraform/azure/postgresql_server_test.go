package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestPostgreSQLServer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "azurerm_resource_group" "example" {
			name     = "example-resources"
			location = "eastus"
		}
		
		resource "azurerm_postgresql_server" "basic_2core" {
			name                = "example-mariadb-server"
			location            = azurerm_resource_group.example.location
			resource_group_name = azurerm_resource_group.example.name
		
			administrator_login          = "fake"
			administrator_login_password = "fake"
		
			sku_name   = "B_Gen5_2"
			storage_mb = 5120
			version    = "9.6"
		
			geo_redundant_backup_enabled  = false
			ssl_enforcement_enabled       = true
		}
		resource "azurerm_postgresql_server" "gp_4core" {
			name                = "example-mariadb-server"
			location            = azurerm_resource_group.example.location
			resource_group_name = azurerm_resource_group.example.name
		
			administrator_login          = "fake"
			administrator_login_password = "fake"
		
			sku_name   = "GP_Gen5_4"
			storage_mb = 4096000
			version    = "9.6"
		
			geo_redundant_backup_enabled  = false
			ssl_enforcement_enabled       = true
		}
		
		resource "azurerm_postgresql_server" "mo_16core" {
			name                = "example-mariadb-server"
			location            = azurerm_resource_group.example.location
			resource_group_name = azurerm_resource_group.example.name
		
			administrator_login          = "fake"
			administrator_login_password = "fake"
		
			sku_name   = "MO_Gen5_16"
			storage_mb = 5120
			version    = "9.6"
		
			geo_redundant_backup_enabled  = false
			ssl_enforcement_enabled       = true
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "azurerm_resource_group.example",
			SkipCheck: true,
		},
		{
			Name: "azurerm_postgresql_server.basic_2core",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (B_Gen5_2)",
					PriceHash:       "01c0bdda146904a0738584748d513ccc-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "9b9812808e5f08d06da81c7aa76a70a3-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(5)),
				},
				{
					Name:             "Additional backup storage",
					PriceHash:        "4077a3c9b2951b50396349b27617d708-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
		{
			Name: "azurerm_postgresql_server.gp_4core",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (GP_Gen5_4)",
					PriceHash:       "086c76267a9b646ad768b00480260569-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "2f6ec013f0c144adfc665854f38a4593-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(4000)),
				},
				{
					Name:             "Additional backup storage",
					PriceHash:        "4077a3c9b2951b50396349b27617d708-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
		{
			Name: "azurerm_postgresql_server.mo_16core",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (MO_Gen5_16)",
					PriceHash:       "6c3ac4a4302d79089dbb83ee5f253079-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "2f6ec013f0c144adfc665854f38a4593-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(5)),
				},
				{
					Name:             "Additional backup storage",
					PriceHash:        "4077a3c9b2951b50396349b27617d708-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestPostgreSQLServer_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "azurerm_resource_group" "example" {
			name     = "example-resources"
			location = "eastus"
		}
		resource "azurerm_postgresql_server" "without_geo" {
			name                = "example-mariadb-server"
			location            = azurerm_resource_group.example.location
			resource_group_name = azurerm_resource_group.example.name
		
			administrator_login          = "fake"
			administrator_login_password = "fake"
		
			sku_name   = "MO_Gen5_16"
			storage_mb = 5120
			version    = "9.6"
		
			geo_redundant_backup_enabled  = false
			ssl_enforcement_enabled       = true
		}
		resource "azurerm_postgresql_server" "with_geo" {
			name                = "example-mariadb-server"
			location            = azurerm_resource_group.example.location
			resource_group_name = azurerm_resource_group.example.name
		
			administrator_login          = "fake"
			administrator_login_password = "fake"
		
			sku_name   = "GP_Gen5_4"
			storage_mb = 4096000
			version    = "9.6"
		
			geo_redundant_backup_enabled  = true
			ssl_enforcement_enabled       = true
		}
		`

	usage := schema.NewUsageMap(map[string]interface{}{
		"azurerm_postgresql_server.without_geo": map[string]interface{}{
			"additional_backup_storage_gb": 2000,
		},
		"azurerm_postgresql_server.with_geo": map[string]interface{}{
			"additional_backup_storage_gb": 3000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "azurerm_resource_group.example",
			SkipCheck: true,
		},
		{
			Name: "azurerm_postgresql_server.without_geo",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (MO_Gen5_16)",
					PriceHash:       "6c3ac4a4302d79089dbb83ee5f253079-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "2f6ec013f0c144adfc665854f38a4593-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(5)),
				},
				{
					Name:             "Additional backup storage",
					PriceHash:        "4077a3c9b2951b50396349b27617d708-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(2000)),
				},
			},
		},
		{
			Name: "azurerm_postgresql_server.with_geo",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (GP_Gen5_4)",
					PriceHash:       "086c76267a9b646ad768b00480260569-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "2f6ec013f0c144adfc665854f38a4593-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(4000)),
				},
				{
					Name:             "Additional backup storage",
					PriceHash:        "eff4b48b0af19750b47fc64108dd71ab-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(3000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
