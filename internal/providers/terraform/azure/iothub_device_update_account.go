package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getIotHubDeviceUpdateAccountRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_iothub_device_update_account",
		CoreRFunc: NewIotHubDeviceUpdateAccount,
	}
}

func NewIotHubDeviceUpdateAccount(d *schema.ResourceData) schema.CoreResource {
	return &azure.IotHubDeviceUpdateAccount{
		Address: d.Address,
		Region:  d.Region,
	}
}
