package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAzureRMLinuxVirtualMachine(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "azurerm_linux_virtual_machine" "standard_f2" {
			name                = "standardlinuxvm"
			resource_group_name = "testrg"
			location            = "uksouth"

			size           = "Standard_F2"
			admin_username = "adminuser"
			admin_password = "T3mp0r4ry!"

			network_interface_ids = [
				"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Network/networkInterfaces/testnic",
			]

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

		resource "azurerm_linux_virtual_machine" "standard_ds2_v2" {
			name                = "standardlinuxvm"
			resource_group_name = "testrg"
			location            = "uksouth"

			size           = "Standard_DS2_v2"
			admin_username = "adminuser"
			admin_password = "T3mp0r4ry!"

			network_interface_ids = [
				"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/testrg/providers/Microsoft.Network/networkInterfaces/testnic",
			]

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

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "azurerm_linux_virtual_machine.standard_f2",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "(Consumption, F2)",
					PriceHash:        "0ef60fde46609c1c899035e94296e51a-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730)),
				},
			},
		},
		{
			Name: "azurerm_linux_virtual_machine.standard_ds2_v2",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "(Consumption, DS2 v2)",
					PriceHash:        "6e2385900ffe7b67c8f24fd791f8142a-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
