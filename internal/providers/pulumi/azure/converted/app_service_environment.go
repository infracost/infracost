package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAppServiceEnvironmentRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_app_service_environment",
		RFunc: NewAppServiceEnvironment,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}
func NewAppServiceEnvironment(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.AppServiceEnvironment{
		Address:     d.Address,
		Region:      d.Region,
		PricingTier: d.Get("pricingTier").String(),
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
