package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getPowerBIEmbeddedRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_powerbi_embedded",
		CoreRFunc: newPowerBIEmbedded,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newPowerBIEmbedded(d *schema.ResourceData) schema.CoreResource {
	region := d.Region
	return &azure.PowerBIEmbedded{
		Address: d.Address,
		Region:  region,
		SKU:     d.Get("sku_name").String(),
	}
}
