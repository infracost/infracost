package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAzureRMAppServicePlan(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
resource "azurerm_resource_group" "example1" {
	name     = "exampleRG1"
	location = "eastus"
  }
  resource "azurerm_app_service_plan" "standard" {
	name                = "api-appserviceplan-pro"
	location            = azurerm_resource_group.example1.location
	resource_group_name = azurerm_resource_group.example1.name
	kind                = "Windows"
	reserved = false
  
	sku {
	  tier = "Standard"
	  size = "S1"
	  capacity = 1
	}
  }
  resource "azurerm_app_service_plan" "standard2" {
	name                = "api-appserviceplan-pro"
	location            = azurerm_resource_group.example1.location
	resource_group_name = azurerm_resource_group.example1.name
	kind                = "Windows"
	reserved = false
  
	sku {
	  tier = "Standard"
	  size = "S1"
	  capacity = 5
	}
  }

  resource "azurerm_app_service_plan" "premium" {
	name                = "api-appserviceplan-pro"
	location            = azurerm_resource_group.example1.location
	resource_group_name = azurerm_resource_group.example1.name
	kind                = "Linux"
	reserved = false
  
	sku {
	  tier = "PremiumContainer"
	  size = "P1v2"
	  capacity = 1
	}
  }
  resource "azurerm_app_service_plan" "premium2" {
	name                = "api-appserviceplan-pro"
	location            = azurerm_resource_group.example1.location
	resource_group_name = azurerm_resource_group.example1.name
	kind                = "Linux"
	reserved = false
  
	sku {
	  tier = "PremiumContainer"
	  size = "P1v2"
	  capacity = 10
	}
  }

  resource "azurerm_app_service_plan" "basic" {
	name                = "api-appserviceplan-pro"
	location            = azurerm_resource_group.example1.location
	resource_group_name = azurerm_resource_group.example1.name
	kind                = "Linux"
	reserved = false
  
	sku {
	  tier = "Basic"
	  size = "B2"
	  capacity = 1
	}
  }
  resource "azurerm_app_service_plan" "basic2" {
	name                = "api-appserviceplan-pro"
	location            = azurerm_resource_group.example1.location
	resource_group_name = azurerm_resource_group.example1.name
	kind                = "Linux"
	reserved = false
  
	sku {
	  tier = "Basic"
	  size = "B2"
	  capacity = 15
	}
  }
`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "azurerm_resource_group.example1",
			SkipCheck: true,
		},
		{
			Name: "azurerm_app_service_plan.standard",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (S1)",
					PriceHash:       "baf738b897c1d3becc742684ce55ed4e-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "azurerm_app_service_plan.standard2",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (S1)",
					PriceHash:       "baf738b897c1d3becc742684ce55ed4e-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(5)),
				},
			},
		},
		{
			Name: "azurerm_app_service_plan.premium",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (P1v2)",
					PriceHash:       "e1e0b7908903429588dbf4ac4d96632d-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "azurerm_app_service_plan.premium2",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (P1v2)",
					PriceHash:       "e1e0b7908903429588dbf4ac4d96632d-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(10)),
				},
			},
		},

		{
			Name: "azurerm_app_service_plan.basic",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (B2)",
					PriceHash:       "b9e04113c366e2a43696a4c13bae496f-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "azurerm_app_service_plan.basic2",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (B2)",
					PriceHash:       "b9e04113c366e2a43696a4c13bae496f-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(15)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
