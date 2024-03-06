package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getVirtualHubRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_virtual_hub",
		CoreRFunc: newVirtualHub,
	}
}

func newVirtualHub(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("region").String()
	sku := "Basic"
	s := d.Get("sku").String()
	if s != "" {
		sku = s
	}

	v := &azure.VirtualHub{
		Address: d.Address,
		Region:  region,
		SKU:     sku,
	}

	return v
}
