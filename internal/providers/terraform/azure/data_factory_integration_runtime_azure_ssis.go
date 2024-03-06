package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getDataFactoryIntegrationRuntimeAzureSSISRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_data_factory_integration_runtime_azure_ssis",
		CoreRFunc: newDataFactoryIntegrationRuntimeAzureSSIS,
	}
}

func newDataFactoryIntegrationRuntimeAzureSSIS(d *schema.ResourceData) schema.CoreResource {
	region := lookupRegion(d, []string{})

	licenseType := d.GetStringOrDefault("license_type", "LicenseIncluded")
	licenseIncluded := strings.EqualFold(licenseType, "LicenseIncluded")

	edition := d.GetStringOrDefault("edition", "Standard")
	enterprise := strings.EqualFold(edition, "Enterprise")

	nodes := d.GetInt64OrDefault("number_of_nodes", 1)

	nodeType := d.Get("node_size").String()
	instanceType := strings.ReplaceAll(nodeType, "Standard_", "")
	instanceType = strings.ReplaceAll(instanceType, "_", " ")

	r := &azure.DataFactoryIntegrationRuntimeAzureSSIS{
		Address:         d.Address,
		Region:          region,
		Enterprise:      enterprise,
		LicenseIncluded: licenseIncluded,
		Instances:       nodes,
		InstanceType:    instanceType,
	}
	return r
}
