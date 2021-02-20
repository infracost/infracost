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
		resource "azurerm_linux_virtual_machine" "standard" {
			name                = "standardlinuxvm"
			resource_group_name = "testrg"
			location            = "UK South"

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
			Name: "azurerm_linux_virtual_machine.standard",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Linux/UNIX usage (Consumption, DS2 v2)",
					PriceHash:        "b728239de79199e8eca1cedc13f48c53-60fc60896424f2f0b576ec5c4e380288",
					HourlyCostCheck:  testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(730)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
