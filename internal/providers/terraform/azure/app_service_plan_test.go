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
		resource "azurerm_app_service_plan" "linux_standard_s3" {
			name                = "api-lin-appserviceplan-std-s3"
			location            = "uksouth"
			resource_group_name = "testrg"
			kind                = "Linux"
			reserved            = true
		
			sku {
				tier = "Standard"
				size = "S3"
			}
		}

		resource "azurerm_app_service_plan" "linux_premium_p1_version_2" {
		    name                = "api-lin-appserviceplan-prem-v2"
			location            = "UK South"
		    resource_group_name = "testrg"
		    kind                = "Linux"
		    reserved            = true
		
		    sku {
		       tier = "Premium"
		       size = "P1 v2"
		    }
		}

		resource "azurerm_app_service_plan" "linux_premium_p1_version_3" {
		    name                = "api-lin-appserviceplan-prem-v3"
			location            = "uksouth"
		    resource_group_name = "testrg"
		    kind                = "Linux"
		    reserved            = true
		
		    sku {
		       tier = "Premium"
		       size = "P1 v3"
		    }
		}

		resource "azurerm_app_service_plan" "windows_premium_container_plan_pc4" {
		    name                = "api-win-appserviceplan-pc4"
			location            = "uksouth"
		    resource_group_name = "testrg"
		    kind                = "Windows"
		    reserved            = false
		
		    sku {
		       tier = "Premium"
		       size = "PC4"
		    }
		}
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "azurerm_app_service_plan.linux_standard_s3",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Consumption usage (Consumption, S3)",
					PriceHash:        "7551c2ba57d451c6ee68825f1ccf7a08-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730)),
				},
			},
		},
		{
			Name: "azurerm_app_service_plan.linux_premium_p1_version_2",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Consumption usage (Consumption, P1 v2)",
					PriceHash:        "2d7c2bd6cdbc3c1b6d9add52a291f176-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730)),
				},
			},
		},
		{
			Name: "azurerm_app_service_plan.linux_premium_p1_version_3",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Consumption usage (Consumption, P1 v3)",
					PriceHash:        "ae4b2eef471a2d364876cb6a2eff735a-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730)),
				},
			},
		},
		{
			Name: "azurerm_app_service_plan.windows_premium_container_plan_pc4",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Consumption usage (Consumption, PC4)",
					PriceHash:        "c1f57f264655d4f05e61b05b06698468-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
