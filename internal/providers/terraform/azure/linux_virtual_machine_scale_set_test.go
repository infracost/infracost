package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAzureRMLinuxVirtualMachineScaleSet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "azurerm_linux_virtual_machine_scale_set" "basic_a2" {
			name                = "basic_a2"
			resource_group_name = "fake_resource_group"
			location            = "eastus"
			instances           = 3

			sku            = "Basic_A2"
			admin_username = "fakeuser"
			admin_password = "fakepass"

			network_interface {
				name    = "example"
				primary = true
		
				ip_configuration {
					name      = "internal"
					primary   = true
					subnet_id = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Network/subnets/fakesubnet"
				}
			}

			os_disk {
				caching              = "ReadWrite"
				storage_account_type = "Standard_LRS"
			}

			source_image_reference {
				publisher = "Canonical"
				offer     = "UbuntuServer"
				sku       = "16.04-LTS"
				version   = "latest"
			}
		}

		resource "azurerm_linux_virtual_machine_scale_set" "basic_a2_usage" {
			name                = "basic_a2"
			resource_group_name = "fake_resource_group"
			location            = "eastus"
			instances           = 3

			sku            = "Basic_A2"
			admin_username = "fakeuser"
			admin_password = "fakepass"

			network_interface {
				name    = "example"
				primary = true
		
				ip_configuration {
					name      = "internal"
					primary   = true
					subnet_id = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Network/subnets/fakesubnet"
				}
			}

			os_disk {
				caching              = "ReadWrite"
				storage_account_type = "Standard_LRS"
			}

			source_image_reference {
				publisher = "Canonical"
				offer     = "UbuntuServer"
				sku       = "16.04-LTS"
				version   = "latest"
			}
		}
	`

	usage := schema.NewUsageMap(map[string]interface{}{
		"azurerm_linux_virtual_machine_scale_set.basic_a2_usage": map[string]interface{}{
			"instances": 4,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "azurerm_linux_virtual_machine_scale_set.basic_a2",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (pay as you go, A2)",
					PriceHash:       "4a86d8e93f2661b5e00bbd43d589f6a9-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(3)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "os_disk",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:             "Storage (S4)",
							PriceHash:        "c972de114a273694e428f2fd1f5fad35-e285791b6e6926c07354b58a33e7ecf4",
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(3)),
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
			Name: "azurerm_linux_virtual_machine_scale_set.basic_a2_usage",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance usage (pay as you go, A2)",
					PriceHash:       "4a86d8e93f2661b5e00bbd43d589f6a9-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(4)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "os_disk",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:             "Storage (S4)",
							PriceHash:        "c972de114a273694e428f2fd1f5fad35-e285791b6e6926c07354b58a33e7ecf4",
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(4)),
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
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
