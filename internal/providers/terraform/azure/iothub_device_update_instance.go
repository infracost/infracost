package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getIotHubDeviceUpdateInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_iothub_device_update_instance",
		CoreRFunc: NewIotHubDeviceUpdateInstance,
	}
}

func NewIotHubDeviceUpdateInstance(d *schema.ResourceData) schema.CoreResource {
	return &azure.IotHubDeviceUpdateInstance{
		Address: d.Address,
		Region:  d.Region,
		Sku:     d.Get("sku").String(), 
	}
}
