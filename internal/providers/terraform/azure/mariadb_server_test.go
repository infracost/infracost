package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestMariaDBServer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "azurerm_resource_group" "example" {
			name     = "example-resources"
			location = "eastus"
		}
		
		resource "azurerm_mariadb_server" "basic_2core" {
			name                = "example-mariadb-server"
			location            = azurerm_resource_group.example.location
			resource_group_name = azurerm_resource_group.example.name
		
			administrator_login          = "fake"
			administrator_login_password = "fake"
		
			sku_name   = "B_Gen5_2"
			storage_mb = 5120
			version    = "10.2"
		
			geo_redundant_backup_enabled  = false
			ssl_enforcement_enabled       = true
		}

		resource "azurerm_mariadb_server" "gp_4core" {
			name                = "example-mariadb-server"
			location            = azurerm_resource_group.example.location
			resource_group_name = azurerm_resource_group.example.name
		
			administrator_login          = "fake"
			administrator_login_password = "fake"
		
			sku_name   = "GP_Gen5_4"
			storage_mb = 4096000
			version    = "10.2"
		
			geo_redundant_backup_enabled  = false
			ssl_enforcement_enabled       = true
		}
		
		resource "azurerm_mariadb_server" "mo_16core" {
			name                = "example-mariadb-server"
			location            = azurerm_resource_group.example.location
			resource_group_name = azurerm_resource_group.example.name
		
			administrator_login          = "fake"
			administrator_login_password = "fake"
		
			sku_name   = "MO_Gen5_16"
			storage_mb = 5120
			version    = "10.3"
		
			geo_redundant_backup_enabled  = false
			ssl_enforcement_enabled       = true
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "azurerm_resource_group.example",
			SkipCheck: true,
		},
		{
			Name: "azurerm_mariadb_server.basic_2core",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (B_Gen5_2)",
					PriceHash:       "6914d5467a1d630659f04aff6ca4d5a6-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "73c878c435ec33946f3cfd3aa7681679-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(5)),
				},
				{
					Name:             "Additional backup storage",
					PriceHash:        "921b28db7dd011a1ad1b1fbe5741a37b-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
		{
			Name: "azurerm_mariadb_server.gp_4core",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (GP_Gen5_4)",
					PriceHash:       "b93e5a8de0d14bb1f2ec965aa1d6b153-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "6de6a33ecfe17c2f4c760f045edf7ae2-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(4000)),
				},
				{
					Name:             "Additional backup storage",
					PriceHash:        "921b28db7dd011a1ad1b1fbe5741a37b-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
		{
			Name: "azurerm_mariadb_server.mo_16core",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (MO_Gen5_16)",
					PriceHash:       "b65ffe3d7afa812a083193075c2d494e-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "6de6a33ecfe17c2f4c760f045edf7ae2-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(5)),
				},
				{
					Name:             "Additional backup storage",
					PriceHash:        "921b28db7dd011a1ad1b1fbe5741a37b-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestMariaDBServer_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "azurerm_resource_group" "example" {
			name     = "example-resources"
			location = "eastus"
		}

		resource "azurerm_mariadb_server" "without_geo" {
			name                = "example-mariadb-server"
			location            = azurerm_resource_group.example.location
			resource_group_name = azurerm_resource_group.example.name
		
			administrator_login          = "fake"
			administrator_login_password = "fake"
		
			sku_name   = "MO_Gen5_16"
			storage_mb = 5120
			version    = "10.3"
		
			geo_redundant_backup_enabled  = false
			ssl_enforcement_enabled       = true
		}

		resource "azurerm_mariadb_server" "with_geo" {
			name                = "example-mariadb-server"
			location            = azurerm_resource_group.example.location
			resource_group_name = azurerm_resource_group.example.name
		
			administrator_login          = "fake"
			administrator_login_password = "fake"
		
			sku_name   = "GP_Gen5_4"
			storage_mb = 4096000
			version    = "10.2"
		
			geo_redundant_backup_enabled  = true
			ssl_enforcement_enabled       = true
		}
		`

	usage := schema.NewUsageMap(map[string]interface{}{
		"azurerm_mariadb_server.without_geo": map[string]interface{}{
			"additional_backup_storage_gb": 2000,
		},
		"azurerm_mariadb_server.with_geo": map[string]interface{}{
			"additional_backup_storage_gb": 3000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "azurerm_resource_group.example",
			SkipCheck: true,
		},
		{
			Name: "azurerm_mariadb_server.without_geo",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (MO_Gen5_16)",
					PriceHash:       "b65ffe3d7afa812a083193075c2d494e-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "6de6a33ecfe17c2f4c760f045edf7ae2-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(5)),
				},
				{
					Name:             "Additional backup storage",
					PriceHash:        "921b28db7dd011a1ad1b1fbe5741a37b-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(2000)),
				},
			},
		},
		{
			Name: "azurerm_mariadb_server.with_geo",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (GP_Gen5_4)",
					PriceHash:       "b93e5a8de0d14bb1f2ec965aa1d6b153-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "6de6a33ecfe17c2f4c760f045edf7ae2-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(4000)),
				},
				{
					Name:             "Additional backup storage",
					PriceHash:        "13a391c49f537e440d031fcd205475f1-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(3000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
