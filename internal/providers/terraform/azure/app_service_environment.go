package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAppServiceEnvironmentRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_app_service_environment",
		CoreRFunc: NewAppServiceEnvironment,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}
func NewAppServiceEnvironment(d *schema.ResourceData) schema.CoreResource {
	r := &azure.AppServiceEnvironment{
		Address:     d.Address,
		Region:      d.Region,
		PricingTier: d.Get("pricing_tier").String(),
	}
	return r
}
