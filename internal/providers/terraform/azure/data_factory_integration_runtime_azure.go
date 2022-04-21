package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getDataFactoryIntegrationRuntimeAzureRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_data_factory_integration_runtime_azure",
		RFunc: newDataFactoryIntegrationRuntimeAzure,
	}
}

func newDataFactoryIntegrationRuntimeAzure(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})

	cores := d.GetInt64OrDefault("core_count", 8)
	compute := d.GetStringOrDefault("compute_type", "General")

	computeType := map[string]string{
		"General":          "general",
		"ComputeOptimized": "compute_optimized",
		"MemoryOptimized":  "memory_optimized",
	}[compute]

	r := &azure.DataFactoryIntegrationRuntimeAzure{
		Address:     d.Address,
		Region:      region,
		Cores:       cores,
		ComputeType: computeType,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
