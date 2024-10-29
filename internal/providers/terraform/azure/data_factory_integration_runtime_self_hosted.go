package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getDataFactoryIntegrationRuntimeSelfHostedRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_data_factory_integration_runtime_self_hosted",
		CoreRFunc: newDataFactoryIntegrationRuntimeSelfHosted,
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

func newDataFactoryIntegrationRuntimeSelfHosted(d *schema.ResourceData) schema.CoreResource {
	r := &azure.DataFactoryIntegrationRuntimeSelfHosted{
		Address: d.Address,
		Region:  d.Region,
	}
	return r
}
