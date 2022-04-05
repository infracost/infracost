package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAzureRMAppIsolatedServicePlanRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_app_service_environment",
		RFunc: NewAppServiceEnvironment,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}
func NewAppServiceEnvironment(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.AppServiceEnvironment{Address: d.Address, Region: lookupRegion(d, []string{"resource_group_name"})}
	if !d.IsEmpty("pricing_tier") {
		r.PricingTier = d.Get("pricing_tier").String()
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
