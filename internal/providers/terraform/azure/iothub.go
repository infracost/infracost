package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getIoTHubRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_iothub",
		CoreRFunc: newIoTHub,
	}
}

func getIoTHubDPSRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_iothub_dps",
		CoreRFunc: newIoTHubDPS,
	}
}

func newIoTHub(d *schema.ResourceData) schema.CoreResource {
	region := d.Region

	sku := d.Get("sku.0.name").String()
	capacity := d.Get("sku.0.capacity").Int()

	r := &azure.IoTHub{
		Address:  d.Address,
		Region:   region,
		Sku:      sku,
		Capacity: capacity,
	}

	return r
}

func newIoTHubDPS(d *schema.ResourceData) schema.CoreResource {
	region := d.Region

	sku := d.Get("sku.0.name").String()

	r := &azure.IoTHubDPS{
		Address: d.Address,
		Region:  region,
		Sku:     sku,
	}

	return r
}
