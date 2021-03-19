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
		resource "azurerm_windows_virtual_machine" "standard_f2" {
			name                = "standardwinvm"
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
				publisher = "MicrosoftWindowsServer"
				offer     = "WindowsServer"
				sku       = "2016-Datacenter"
				version   = "latest"
			  }
		}
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "azurerm_windows_virtual_machine.standard_f2",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "(Consumption, F2)",
					PriceHash:        "744969fce4ad19dd7c8a99d360619c13-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
