package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getDataFactoryIntegrationRuntimeAzureRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_data_factory_integration_runtime_azure",
		CoreRFunc: newDataFactoryIntegrationRuntimeAzure,
		ReferenceAttributes: []string{
			"data_factory_id",
			"data_factory_name",
			"resource_group_name",
		},
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			region := lookupRegion(d, []string{"resource_group_name", "data_factory_id", "data_factory_name"})

			dataFactoryIdRefs := d.References("data_factory_id")
			if region == "" && len(dataFactoryIdRefs) > 0 {
				region = lookupRegion(dataFactoryIdRefs[0], []string{"resource_group_name"})
			}

			// Old provider versions < 3 can reference data_factory_name
			dataFactoryNameRefs := d.References("data_factory_name")
			if region == "" && len(dataFactoryNameRefs) > 0 {
				region = lookupRegion(dataFactoryNameRefs[0], []string{"resource_group_name"})
			}

			return region
		},
	}
}

func newDataFactoryIntegrationRuntimeAzure(d *schema.ResourceData) schema.CoreResource {
	cores := d.GetInt64OrDefault("core_count", 8)
	compute := d.GetStringOrDefault("compute_type", "General")

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
