package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAppServicePlanRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_app_service_plan",
		CoreRFunc: NewAppServicePlan,
	}
}
func NewAppServicePlan(d *schema.ResourceData) schema.CoreResource {
	r := &azure.AppServicePlan{
		Address:     d.Address,
		Region:      d.Region,
		SKUSize:     d.Get("sku.0.size").String(),
		SKUCapacity: d.Get("sku.0.capacity").Int(),
		Kind:        d.Get("kind").String(),
		IsDevTest:   d.ProjectMetadata["isProduction"] == "false",
	}
	return r
}
