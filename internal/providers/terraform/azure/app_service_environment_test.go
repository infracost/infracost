package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAzureRMAppIsolatedServicePlan(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	tf := `

resource "azurerm_resource_group" "example1" {
	name     = "exampleRG1"
	location = "eastus"
  }
  resource "azurerm_subnet" "ase" {
	name                 = "asesubnet"
	resource_group_name  = azurerm_resource_group.example1.name
	virtual_network_name = azurerm_virtual_network.example1.name
	address_prefixes     = ["10.0.1.0/24"]
  }
  resource "azurerm_virtual_network" "example1" {
	name                = "example-vnet1"
	location            = azurerm_resource_group.example1.location
	resource_group_name = azurerm_resource_group.example1.name
	address_space       = ["10.0.0.0/16"]
  }

resource "azurerm_app_service_environment" "example" {
	name                         = "example-ase"
	subnet_id                    = azurerm_subnet.ase.id
	pricing_tier                 = "I1"
	front_end_scale_factor       = 10
	internal_load_balancing_mode = "Web, Publishing"
	allowed_user_ip_cidrs        = ["11.22.33.44/32", "55.66.77.0/24"]
	resource_group_name          =  azurerm_resource_group.example1.name
  
   
	cluster_setting {
	  name  = "DisableTls1.0"
	  value = "1"
	}
  }
  resource "azurerm_app_service_environment" "example2" {
	name                         = "example-ase"
	subnet_id                    = azurerm_subnet.ase.id
	pricing_tier                 = "I2"
	front_end_scale_factor       = 10
	internal_load_balancing_mode = "Web, Publishing"
	allowed_user_ip_cidrs        = ["11.22.33.44/32", "55.66.77.0/24"]
	resource_group_name          =  azurerm_resource_group.example1.name
  
   
	cluster_setting {
	  name  = "DisableTls1.0"
	  value = "1"
	}
  }
`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "azurerm_resource_group.example1",
			SkipCheck: true,
		},
		{
			Name:      "azurerm_subnet.ase",
			SkipCheck: true,
		},
		{
			Name: "azurerm_app_service_environment.example",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Stamp fee",
					PriceHash:       "b74a7d0a1aca40309bb9ad96939232c7-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "Instance usage (I1)",
					PriceHash:       "e24d279fdaf82f6735d867987e90c11a-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "azurerm_app_service_environment.example2",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Stamp fee",
					PriceHash:       "b74a7d0a1aca40309bb9ad96939232c7-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "Instance usage (I2)",
					PriceHash:       "28a74220fd3d2c3872dcf0d01f1dedca-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
func TestAzureRMAppIsolatedServicePlan_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	tf := `

resource "azurerm_resource_group" "example1" {
	name     = "exampleRG1"
	location = "eastus"
  }
  resource "azurerm_subnet" "ase" {
	name                 = "asesubnet"
	resource_group_name  = azurerm_resource_group.example1.name
	virtual_network_name = azurerm_virtual_network.example1.name
	address_prefixes     = ["10.0.1.0/24"]
  }
  resource "azurerm_virtual_network" "example1" {
	name                = "example-vnet1"
	location            = azurerm_resource_group.example1.location
	resource_group_name = azurerm_resource_group.example1.name
	address_space       = ["10.0.0.0/16"]
  }

resource "azurerm_app_service_environment" "example" {
	name                         = "example-ase"
	subnet_id                    = azurerm_subnet.ase.id
	pricing_tier                 = "I1"
	front_end_scale_factor       = 10
	internal_load_balancing_mode = "Web, Publishing"
	allowed_user_ip_cidrs        = ["11.22.33.44/32", "55.66.77.0/24"]
	resource_group_name          =  azurerm_resource_group.example1.name
  
   
	cluster_setting {
	  name  = "DisableTls1.0"
	  value = "1"
	}
  }
  resource "azurerm_app_service_environment" "example2" {
	name                         = "example-ase"
	subnet_id                    = azurerm_subnet.ase.id
	pricing_tier                 = "I2"
	front_end_scale_factor       = 10
	internal_load_balancing_mode = "Web, Publishing"
	allowed_user_ip_cidrs        = ["11.22.33.44/32", "55.66.77.0/24"]
	resource_group_name          =  azurerm_resource_group.example1.name
  
   
	cluster_setting {
	  name  = "DisableTls1.0"
	  value = "1"
	}
  }
`
	usage := schema.NewUsageMap(map[string]interface{}{
		"azurerm_app_service_environment.example": map[string]interface{}{
			"operating_system": "windows",
		},
		"azurerm_app_service_environment.example2": map[string]interface{}{
			"operating_system": "windows",
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "azurerm_resource_group.example1",
			SkipCheck: true,
		},
		{
			Name:      "azurerm_subnet.ase",
			SkipCheck: true,
		},
		{
			Name: "azurerm_app_service_environment.example",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Stamp fee",
					PriceHash:       "c73f0ef5302d9ba9a4b0c8d6fed8f237-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "Instance usage (I1)",
					PriceHash:       "59e240e88eb05464adb6515f2d67d618-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "azurerm_app_service_environment.example2",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Stamp fee",
					PriceHash:       "c73f0ef5302d9ba9a4b0c8d6fed8f237-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "Instance usage (I2)",
					PriceHash:       "c1dba1507c40baabb9269bac266f9b90-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
