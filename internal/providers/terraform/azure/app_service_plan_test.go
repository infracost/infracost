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
			location            = "UK South"
			resource_group_name = "testrg"
			kind                = "Linux"
			reserved            = false
		
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
		    reserved            = false
		
		    sku {
		       tier = "Premium"
		       size = "P1 v2"
		    }
		}

		resource "azurerm_app_service_plan" "linux_premium_p1_version_3" {
		    name                = "api-lin-appserviceplan-prem-v3"
		    location            = "UK South"
		    resource_group_name = "testrg"
		    kind                = "Linux"
		    reserved            = false
		
		    sku {
		       tier = "Premium"
		       size = "P1 v3"
		    }
		}

		resource "azurerm_app_service_plan" "windows_premium_container_plan_pc4" {
		    name                = "api-win-appserviceplan-pc4"
		    location            = "UK South"
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
					Name:             "Computing usage (Consumption, S3)",
					PriceHash:        "d2ca2764fbf2fa06736cf31a8c90ce27-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730)),
				},
			},
		},
		{
			Name: "azurerm_app_service_plan.linux_premium_p1_version_2",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Computing usage (Consumption, P1 v2)",
					PriceHash:        "b10671e31a113e54bb32c30152591531-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730)),
				},
			},
		},
		{
			Name: "azurerm_app_service_plan.linux_premium_p1_version_3",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Computing usage (Consumption, P1 v3)",
					PriceHash:        "3e17d3c1dc457ac52a660779547a3ebd-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730)),
				},
			},
		},
		{
			Name: "azurerm_app_service_plan.windows_premium_container_plan_pc4",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Computing usage (Consumption, PC4)",
					PriceHash:        "f83bca62167d4fafbe44390c3d63eb0e-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
