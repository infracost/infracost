package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAzureRMWindowsVirtualMachine(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "azurerm_windows_virtual_machine" "basic_a2" {
			name                = "basic_a2"
			resource_group_name = "fake_resource_group"
			location            = "eastus"

			size           = "Basic_A2"
			admin_username = "fakeuser"
			admin_password = "fakepass"

			network_interface_ids = [
				"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Network/networkInterfaces/fakenic",
			]

			os_disk {
				caching              = "ReadWrite"
				storage_account_type = "Standard_LRS"
			}

			source_image_reference {
				publisher = "MicrosoftWindowsServer"
				offer     = "WindowsServer"
				sku       = "2016-Datacenter"
				version   = "latest"
			}
		}

		resource "azurerm_windows_virtual_machine" "standard_f2_premium_disk" {
			name                = "standard_f2"
			resource_group_name = "fake_resource_group"
			location            = "eastus"

			size           = "Standard_F2"
			admin_username = "fakeuser"
			admin_password = "fakepass"

			network_interface_ids = [
				"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Network/networkInterfaces/fakenic",
			]

			os_disk {
				caching              = "ReadWrite"
				storage_account_type = "Premium_LRS"
			}

			source_image_reference {
				publisher = "MicrosoftWindowsServer"
				offer     = "WindowsServer"
				sku       = "2016-Datacenter"
				version   = "latest"
			}
		}

		resource "azurerm_windows_virtual_machine" "standard_a2_v2_custom_disk" {
			name                = "standard_a2_v2_custom_disk"
			resource_group_name = "fake_resource_group"
			location            = "eastus"

			size           = "Standard_A2_v2"
			admin_username = "fakeuser"
			admin_password = "fakepass"

			network_interface_ids = [
				"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Network/networkInterfaces/fakenic",
			]

			os_disk {
				caching              = "ReadWrite"
				storage_account_type = "StandardSSD_LRS"
				disk_size_gb         = 1000
			}

			source_image_reference {
				publisher = "MicrosoftWindowsServer"
				offer     = "WindowsServer"
				sku       = "2016-Datacenter"
				version   = "latest"
			}
		}

		resource "azurerm_windows_virtual_machine" "standard_d2_v4_hybrid_benefit" {
			name                = "standard_a2_v2_custom_disk"
			resource_group_name = "fake_resource_group"
			location            = "eastus"

			size           = "Standard_D2_v4"
			admin_username = "fakeuser"
			admin_password = "fakepass"

			license_type = "Windows_Server"

			network_interface_ids = [
				"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Network/networkInterfaces/fakenic",
			]

			os_disk {
				caching              = "ReadWrite"
				storage_account_type = "StandardSSD_LRS"
				disk_size_gb         = 1000
			}

			source_image_reference {
				publisher = "MicrosoftWindowsServer"
				offer     = "WindowsServer"
				sku       = "2016-Datacenter"
				version   = "latest"
			}
		}
		
		resource "azurerm_windows_virtual_machine" "standard_a2_ultra_enabled" {
			name                = "standard_a2_ultra_enabled"
			resource_group_name = "fake_resource_group"
			location            = "eastus"

			size           = "Standard_A2_v2"
			admin_username = "fakeuser"
			admin_password = "fakepass"

			network_interface_ids = [
				"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Network/networkInterfaces/fakenic",
			]

			os_disk {
				caching              = "ReadWrite"
				storage_account_type = "StandardSSD_LRS"
			}
			
			additional_capabilities {
				ultra_ssd_enabled = true
			}

			source_image_reference {
				publisher = "MicrosoftWindowsServer"
				offer     = "WindowsServer"
				sku       = "2016-Datacenter"
				version   = "latest"
			}
		}
	`

	usage := schema.NewUsageMap(map[string]interface{}{
		"azurerm_windows_virtual_machine.standard_a2_v2_custom_disk": map[string]interface{}{
			"os_disk.monthly_disk_operations": 20000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "azurerm_windows_virtual_machine.basic_a2",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (pay as you go, A2)",
					PriceHash:       "571360f185f641494606f9489a2f7141-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "os_disk",
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
			},
		},
		{
			Name: "azurerm_windows_virtual_machine.standard_f2_premium_disk",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (pay as you go, F2)",
					PriceHash:       "c9c93d374aed19384afe2b52c606adbd-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "os_disk",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:             "Storage (P4)",
							PriceHash:        "6b63e474135f6a1157a2f348bb4fd899-e285791b6e6926c07354b58a33e7ecf4",
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
						},
					},
				},
			},
		},
		{
			Name: "azurerm_windows_virtual_machine.standard_a2_v2_custom_disk",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (pay as you go, A2 v2)",
					PriceHash:       "e2c93467baafc09fd80cea03c9c5c324-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "os_disk",
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
			},
		},
		{
			Name: "azurerm_windows_virtual_machine.standard_d2_v4_hybrid_benefit",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (hybrid benefit, D2 v4)",
					PriceHash:       "188638135dbae0493714883d684fc65a-cfd8f218a181c17d03a5e84e38767fcc",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "os_disk",
					SkipCheck: true,
				},
			},
		},
		{
			Name: "azurerm_windows_virtual_machine.standard_a2_ultra_enabled",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:      "Instance usage (pay as you go, A2 v2)",
					SkipCheck: true,
				},
				{
					Name:             "Ultra disk reservation (if unattached)",
					PriceHash:        "cdb969dd2d88d468c7842c5a42c07050-c1baecc1e3a89596af672fd42fe001bf",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "os_disk",
					SkipCheck: true,
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
