package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAzureRMVirtualHubRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_virtual_hub",
		RFunc: newVirtualHub,
	}
}

func newVirtualHub(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
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
	v.PopulateUsage(u)

	return v.BuildResource()
}
