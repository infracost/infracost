package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getDataFactoryIntegrationRuntimeAzureSSISRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_data_factory_integration_runtime_azure_ssis",
		RFunc: newDataFactoryIntegrationRuntimeAzureSSIS,
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

			// Old provider versions <3 can reference data_factory_name
			dataFactoryNameRefs := d.References("dataFactoryName")
			if region == "" && len(dataFactoryNameRefs) > 0 {
				region = lookupRegion(dataFactoryNameRefs[0], []string{"resource_group_name"})
			}

			return region
		},
	}
}

func newDataFactoryIntegrationRuntimeAzureSSIS(d *schema.ResourceData) schema.CoreResource {
	licenseType := d.GetStringOrDefault("licenseType", "LicenseIncluded")
	licenseIncluded := strings.EqualFold(licenseType, "LicenseIncluded")

	edition := d.GetStringOrDefault("edition", "Standard")
	enterprise := strings.EqualFold(edition, "Enterprise")

	nodes := d.GetInt64OrDefault("numberOfNodes", 1)

	nodeType := d.Get("nodeSize").String()
	instanceType := strings.ReplaceAll(nodeType, "Standard_", "")
	instanceType = strings.ReplaceAll(instanceType, "_", " ")

	r := &azure.DataFactoryIntegrationRuntimeAzureSSIS{
		Address:         d.Address,
		Region:          d.Region,
		Enterprise:      enterprise,
		LicenseIncluded: licenseIncluded,
		Instances:       nodes,
		InstanceType:    instanceType,
	}
	return r
}
