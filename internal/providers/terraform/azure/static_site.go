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

// azurerm_static_web_app is the renamed successor to azurerm_static_site;
// schema is the same so we reuse the same constructor.
func getStaticWebAppRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_static_web_app",
		CoreRFunc: newStaticSite,
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			return lookupRegion(d, []string{"resource_group_name"})
		},
	}
}

func newStaticSite(d *schema.ResourceData) schema.CoreResource {
	sku := "Free"
	if !d.IsEmpty("sku_tier") {
		sku = d.Get("sku_tier").String()
	}

	return &azure.StaticSite{
		Address: d.Address,
		Region:  d.Region,
		SKU:     sku,
	}
}
