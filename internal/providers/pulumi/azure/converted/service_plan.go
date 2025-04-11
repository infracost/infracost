package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getServicePlanRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name: "azurerm_service_plan",
		RFunc: func(d *schema.ResourceData) schema.CoreResource {
			return &azure.ServicePlan{
				Address:     d.Address,
				Region:      d.Region,
				SKUName:     d.Get("skuName").String(),
				WorkerCount: d.GetInt64OrDefault("workerCount", 1),
				OSType:      d.Get("osType").String(),
				IsDevTest:   d.ProjectMetadata["isProduction"] == "false",
			}
		},
	}
}
