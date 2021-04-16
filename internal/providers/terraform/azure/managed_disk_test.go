package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAzureRMManagedDisk(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "azurerm_managed_disk" "standard" {
			name                 = "standard"
			resource_group_name  = "fake_resource_group"
			location             = "eastus"

			create_option        = "Empty"
			storage_account_type = "Standard_LRS"
		}

		resource "azurerm_managed_disk" "premium" {
			name                 = "premium"
			resource_group_name  = "fake_resource_group"
			location             = "eastus"

			create_option        = "Empty"
			storage_account_type = "Premium_LRS"
		}

		resource "azurerm_managed_disk" "custom_size_ssd" {
			name                 = "custom_size_ssd"
			resource_group_name  = "fake_resource_group"
			location             = "eastus"

			create_option        = "Empty"
			storage_account_type = "StandardSSD_LRS"
			disk_size_gb         = 1000
		}

		resource "azurerm_managed_disk" "ultra" {
			name                 = "ultra"
			resource_group_name  = "fake_resource_group"
			location             = "eastus"

			create_option        = "Empty"
			storage_account_type = "UltraSSD_LRS"
			disk_size_gb         = 2000
			disk_iops_read_write = 4000
			disk_mbps_read_write = 20
		}
	`

	usage := schema.NewUsageMap(map[string]interface{}{
		"azurerm_managed_disk.custom_size_ssd": map[string]interface{}{
			"monthly_disk_operations": 20000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "azurerm_managed_disk.standard",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage (S4)",
					PriceHash:        "c972de114a273694e428f2fd1f5fad35-e285791b6e6926c07354b58a33e7ecf4",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Disk operations",
					PriceHash:        "bde05feab07ea46abc6317ffd45fca54-49c37505210dfd1c98233191a87608bd",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
		{
			Name: "azurerm_managed_disk.premium",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage (P4)",
					PriceHash:        "6b63e474135f6a1157a2f348bb4fd899-e285791b6e6926c07354b58a33e7ecf4",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "azurerm_managed_disk.custom_size_ssd",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Storage (E30)",
					PriceHash:        "c481ed02851a82921c43001f551d5759-e285791b6e6926c07354b58a33e7ecf4",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Disk operations",
					PriceHash:        "ed218c1aa21867f98ddaf9e259dc8eb8-49c37505210dfd1c98233191a87608bd",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(2)),
				},
			},
		},
		{
			Name: "azurerm_managed_disk.ultra",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Storage (ultra, 2048 GiB)",
					PriceHash:       "e481f6bc40f35eb04b1557d8a75e9e7b-79cf848d5b25ab30487591294a219c38",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(2048)),
				},
				{
					Name:            "Provisioned IOPS",
					PriceHash:       "71471e4795a660274782a8f5273cb172-c1baecc1e3a89596af672fd42fe001bf",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(4000)),
				},
				{
					Name:            "Throughput",
					PriceHash:       "6bc8532a4b4dce4a0f887edd4abeba25-c1baecc1e3a89596af672fd42fe001bf",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
