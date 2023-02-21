package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getPowerBIEmbeddedRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_powerbi_embedded",
		RFunc: newPowerBIEmbedded,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newPowerBIEmbedded(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{"resource_group_name"})
	r := &azure.PowerBIEmbedded{
		Address: d.Address,
		Region:  region,
		SKU:     d.Get("sku_name").String(),
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
