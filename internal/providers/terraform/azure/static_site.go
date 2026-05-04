package azure

import (
	"strings"

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
	// sku_tier and sku_size are equivalent on the schema. Take whichever is
	// non-Free so a config that sets only sku_size still picks up the paid
	// tier even when terraform plan has filled the other in with the "Free"
	// default.
	sku := "Free"
	for _, attr := range []string{"sku_tier", "sku_size"} {
		v := d.Get(attr).String()
		if v != "" && !strings.EqualFold(v, "Free") {
			sku = v
			break
		}
	}

	return &azure.StaticSite{
		Address: d.Address,
		Region:  d.Region,
		SKU:     sku,
	}
}
