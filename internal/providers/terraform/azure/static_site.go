package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getStaticSiteRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_static_site",
		CoreRFunc: newStaticSite,
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			return lookupRegion(d, []string{"resource_group_name"})
		},
	}
}

func newStaticSite(d *schema.ResourceData) schema.CoreResource {
	region := lookupRegion(d, []string{"resource_group_name"})

	sku := "Free"
	if d.Get("sku_tier").String() != "" {
		sku = d.Get("sku_tier").String()
	}

	return &azure.StaticSite{
		Address: d.Address,
		Region:  region,
		SKU:     sku,
	}
} 