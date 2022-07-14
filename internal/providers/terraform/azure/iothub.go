package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getIoTHubRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_iothub",
		RFunc: newIoTHub,
	}
}

func getIoTHubDPSRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_iothub_dps",
		RFunc: newIoTHubDPS,
	}
}

func newIoTHub(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})

	sku := d.Get("sku.0.name").String()
	capacity := d.Get("sku.0.capacity").Int()

	r := &azure.IoTHub{
		Address:  d.Address,
		Region:   region,
		Sku:      sku,
		Capacity: capacity,
	}

	r.PopulateUsage(u)

	return r.BuildResource()
}

func newIoTHubDPS(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})

	sku := d.Get("sku.0.name").String()

	r := &azure.IoTHubDPS{
		Address: d.Address,
		Region:  region,
		Sku:     sku,
	}

	r.PopulateUsage(u)

	return r.BuildResource()
}
