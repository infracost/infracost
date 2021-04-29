package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestMSSQLDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "azurerm_resource_group" "example" {
		name     = "example-resources"
		location = "eastus"
	}
	
	resource "azurerm_sql_server" "example" {
		name                         = "example-sqlserver"
		resource_group_name          = azurerm_resource_group.example.name
		location                     = "eastus"
		version                      = "12.0"
		administrator_login          = "fake"
		administrator_login_password = "fake"
	}
	
	resource "azurerm_mssql_database" "general_purpose_gen" {
		name           = "acctest-db-d"
		server_id      = azurerm_sql_server.example.id
		sku_name       = "GP_Gen5_4"
	}
	resource "azurerm_mssql_database" "business_critical_gen" {
		name           = "acctest-db-d"
		server_id      = azurerm_sql_server.example.id
		sku_name       = "BC_Gen5_8"
		max_size_gb    = 10
	}
	resource "azurerm_mssql_database" "business_critical_m" {
		name           = "acctest-db-d"
		server_id      = azurerm_sql_server.example.id
		sku_name       = "BC_M_8"
		max_size_gb    = 50
	}
	resource "azurerm_mssql_database" "hyperscale_gen" {
		name           = "acctest-db-d"
		server_id      = azurerm_sql_server.example.id
		sku_name       = "HS_Gen5_2"
		max_size_gb    = 100
	}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "azurerm_resource_group.example",
			SkipCheck: true,
		},
		{
			Name:      "azurerm_sql_server.example",
			SkipCheck: true,
		},
		{
			Name: "azurerm_mssql_database.general_purpose_gen",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (provisioned, GP_Gen5_4)",
					PriceHash:       "f42898ff48d81acaf7b657aacaf277db-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "7614d4c707b678cc053a4e75265fdfee-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(5)),
				},
				{
					Name:             "Long-term retention",
					PriceHash:        "1fd081640191a4301b4354155c39bbee-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
		{
			Name: "azurerm_mssql_database.business_critical_gen",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (provisioned, BC_Gen5_8)",
					PriceHash:       "8ad0fc701dd661618992c51d38418cd5-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "32f0514bc8bade7e28ef38328604cbdc-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10)),
				},
				{
					Name:             "Long-term retention",
					PriceHash:        "1fd081640191a4301b4354155c39bbee-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
		{
			Name: "azurerm_mssql_database.business_critical_m",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (provisioned, BC_M_8)",
					PriceHash:       "6d355ec153d614628a3847666d995f1a-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "32f0514bc8bade7e28ef38328604cbdc-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(50)),
				},
				{
					Name:             "Long-term retention",
					PriceHash:        "1fd081640191a4301b4354155c39bbee-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
		{
			Name: "azurerm_mssql_database.hyperscale_gen",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (provisioned, HS_Gen5_2)",
					PriceHash:       "49a7ab0d61cb0e246bb4a1db18e124bb-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "34bb4ebb9f806ef9b11c951a2dfa2d78-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(100)),
				},
				{
					Name:             "Read replicas",
					PriceHash:        "49a7ab0d61cb0e246bb4a1db18e124bb-60fc60896424f2f0b576ec5c4e380288",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestMSSQLDatabase_HS_with_replicas(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "azurerm_resource_group" "example" {
		name     = "example-resources"
		location = "eastus"
	}
	
	resource "azurerm_sql_server" "example" {
		name                         = "example-sqlserver"
		resource_group_name          = azurerm_resource_group.example.name
		location                     = "eastus"
		version                      = "12.0"
		administrator_login          = "fake"
		administrator_login_password = "fake"
	}
	resource "azurerm_mssql_database" "hyperscale_gen" {
		name           = "acctest-db-d"
		server_id      = azurerm_sql_server.example.id
		sku_name       = "HS_Gen5_2"
		read_replica_count = 2
	}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "azurerm_resource_group.example",
			SkipCheck: true,
		},
		{
			Name:      "azurerm_sql_server.example",
			SkipCheck: true,
		},
		{
			Name: "azurerm_mssql_database.hyperscale_gen",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (provisioned, HS_Gen5_2)",
					PriceHash:       "49a7ab0d61cb0e246bb4a1db18e124bb-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "34bb4ebb9f806ef9b11c951a2dfa2d78-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(5)),
				},
				{
					Name:            "Read replicas",
					PriceHash:       "49a7ab0d61cb0e246bb4a1db18e124bb-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(2)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestMSSQLDatabase_with_license(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "azurerm_resource_group" "example" {
		name     = "example-resources"
		location = "eastus"
	}
	
	resource "azurerm_sql_server" "example" {
		name                         = "example-sqlserver"
		resource_group_name          = azurerm_resource_group.example.name
		location                     = "eastus"
		version                      = "12.0"
		administrator_login          = "fake"
		administrator_login_password = "fake"
	}
	resource "azurerm_mssql_database" "general_purpose_gen" {
		name           = "acctest-db-d"
		server_id      = azurerm_sql_server.example.id
		sku_name       = "GP_Gen5_4"
		license_type   = "LicenseIncluded"
	}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "azurerm_resource_group.example",
			SkipCheck: true,
		},
		{
			Name:      "azurerm_sql_server.example",
			SkipCheck: true,
		},
		{
			Name: "azurerm_mssql_database.general_purpose_gen",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (provisioned, GP_Gen5_4)",
					PriceHash:       "f42898ff48d81acaf7b657aacaf277db-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "SQL license",
					PriceHash:       "884778bbd38c482d0fb8c655195422e7-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(4)),
				},
				{
					Name:             "Storage",
					PriceHash:        "7614d4c707b678cc053a4e75265fdfee-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(5)),
				},
				{
					Name:             "Long-term retention",
					PriceHash:        "1fd081640191a4301b4354155c39bbee-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestMSSQLDatabase_zone_redundant(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "azurerm_resource_group" "example" {
		name     = "example-resources"
		location = "eastus"
	}
	
	resource "azurerm_sql_server" "example" {
		name                         = "example-sqlserver"
		resource_group_name          = azurerm_resource_group.example.name
		location                     = "eastus"
		version                      = "12.0"
		administrator_login          = "fake"
		administrator_login_password = "fake"
	}
	resource "azurerm_mssql_database" "general_purpose_gen" {
		name           = "acctest-db-d"
		server_id      = azurerm_sql_server.example.id
		sku_name       = "GP_Gen5_4"
		zone_redundant = true
	}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "azurerm_resource_group.example",
			SkipCheck: true,
		},
		{
			Name:      "azurerm_sql_server.example",
			SkipCheck: true,
		},
		{
			Name: "azurerm_mssql_database.general_purpose_gen",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (provisioned, GP_Gen5_4)",
					PriceHash:       "ebdf927dde026ddb10d829159f37543c-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "1336a17109124a6daeef5225cda70e88-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(5)),
				},
				{
					Name:             "Long-term retention",
					PriceHash:        "1fd081640191a4301b4354155c39bbee-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestMSSQLDatabase_serverless(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "azurerm_resource_group" "example" {
		name     = "example-resources"
		location = "eastus"
	}
	
	resource "azurerm_sql_server" "example" {
		name                         = "example-sqlserver"
		resource_group_name          = azurerm_resource_group.example.name
		location                     = "eastus"
		version                      = "12.0"
		administrator_login          = "fake"
		administrator_login_password = "fake"
	}
	
	resource "azurerm_mssql_database" "serverless" {
		name           = "acctest-db-d"
		server_id      = azurerm_sql_server.example.id
		sku_name       = "GP_S_Gen5_4"
	}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"azurerm_mssql_database.serverless": map[string]interface{}{
			"monthly_vcore_hours": 500,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "azurerm_resource_group.example",
			SkipCheck: true,
		},
		{
			Name:      "azurerm_sql_server.example",
			SkipCheck: true,
		},
		{
			Name: "azurerm_mssql_database.serverless",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Compute (serverless, GP_S_Gen5_4)",
					PriceHash:        "02cae7437956531c79530f317b7f58ce-60fc60896424f2f0b576ec5c4e380288",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(500)),
				},
				{
					Name:             "Storage",
					PriceHash:        "7614d4c707b678cc053a4e75265fdfee-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(5)),
				},
				{
					Name:             "Long-term retention",
					PriceHash:        "1fd081640191a4301b4354155c39bbee-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}

func TestMSSQLDatabase_LTR(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "azurerm_resource_group" "example" {
		name     = "example-resources"
		location = "eastus"
	}
	
	resource "azurerm_sql_server" "example" {
		name                         = "example-sqlserver"
		resource_group_name          = azurerm_resource_group.example.name
		location                     = "eastus"
		version                      = "12.0"
		administrator_login          = "fake"
		administrator_login_password = "fake"
	}
	
	resource "azurerm_mssql_database" "my_db" {
		name           = "acctest-db-d"
		server_id      = azurerm_sql_server.example.id
		sku_name       = "GP_Gen5_4"
	}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"azurerm_mssql_database.my_db": map[string]interface{}{
			"long_term_retention_storage_gb": 1000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "azurerm_resource_group.example",
			SkipCheck: true,
		},
		{
			Name:      "azurerm_sql_server.example",
			SkipCheck: true,
		},
		{
			Name: "azurerm_mssql_database.my_db",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (provisioned, GP_Gen5_4)",
					PriceHash:       "f42898ff48d81acaf7b657aacaf277db-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Storage",
					PriceHash:        "7614d4c707b678cc053a4e75265fdfee-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(5)),
				},
				{
					Name:             "Long-term retention",
					PriceHash:        "1fd081640191a4301b4354155c39bbee-ea8c44e23e41502dcee5033e136055b6",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
