package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getDataFactoryIntegrationRuntimeAzureRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_data_factory_integration_runtime_azure",
		RFunc: newDataFactoryIntegrationRuntimeAzure,
		ReferenceAttributes: []string{
			"data_factory_id",
			"data_factory_name",
			"resource_group_name",
		},
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			region := lookupRegion(d, []string{"resource_group_name", "data_factory_id", "data_factory_name"})

			dataFactoryIdRefs := d.References("dataFactoryId")
			if region == "" && len(dataFactoryIdRefs) > 0 {
				region = lookupRegion(dataFactoryIdRefs[0], []string{"resource_group_name"})
			}

			// Old provider versions < 3 can reference data_factory_name
			dataFactoryNameRefs := d.References("dataFactoryName")
			if region == "" && len(dataFactoryNameRefs) > 0 {
				region = lookupRegion(dataFactoryNameRefs[0], []string{"resource_group_name"})
			}

			return region
		},
	}
}

func newDataFactoryIntegrationRuntimeAzure(d *schema.ResourceData) schema.CoreResource {
	cores := d.GetInt64OrDefault("coreCount", 8)
	compute := d.GetStringOrDefault("computeType", "General")

	computeType := map[string]string{
		"General":          "general",
		"ComputeOptimized": "compute_optimized",
		"MemoryOptimized":  "memory_optimized",
	}[compute]

	r := &azure.DataFactoryIntegrationRuntimeAzure{
		Address:     d.Address,
		Region:      d.Region,
		Cores:       cores,
		ComputeType: computeType,
	}
	return r
}
