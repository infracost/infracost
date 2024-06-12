package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getLogicAppStandardRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_logic_app_standard",
		CoreRFunc: newLogicAppStandard,
		ReferenceAttributes: []string{
			"resource_group_name",
			"app_service_plan_id",
		},
	}
}

func newLogicAppStandard(d *schema.ResourceData) schema.CoreResource {
	region := d.Region

	var sku *string
	appServicePlans := d.References("app_service_plan_id")
	if len(appServicePlans) > 0 {
		sku = strPtr(appServicePlans[0].Get("sku.0.size").String())
	}

	return &azure.LogicAppStandard{
		Address: d.Address,
		Region:  region,
		SKU:     sku,
	}
}
