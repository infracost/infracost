package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getServicePlanRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_service_plan",
		RFunc: NewServicePlan,
	}
}
func NewServicePlan(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.ServicePlan{
		Address:     d.Address,
		Region:      lookupRegion(d, []string{}),
		SKUName:     d.Get("sku_name").String(),
		WorkerCount: d.GetInt64OrDefault("worker_count", 1),
		OSType:      d.Get("os_type").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
