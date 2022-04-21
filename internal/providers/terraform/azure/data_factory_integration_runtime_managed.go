package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

// This resource is superseded by azurerm_data_factory_integration_runtime_azure_ssis
// in Terraform. Their instance types look the same, but the pricing page
// additionally mentions other operations for managed runtime.
func getDataFactoryIntegrationRuntimeManagedRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_data_factory_integration_runtime_managed",
		RFunc: newDataFactoryIntegrationRuntimeManaged,
	}
}

func newDataFactoryIntegrationRuntimeManaged(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})

	licenseType := d.GetStringOrDefault("license_type", "LicenseIncluded")
	licenseIncluded := strings.EqualFold(licenseType, "LicenseIncluded")

	edition := d.GetStringOrDefault("edition", "Standard")
	enterprise := strings.EqualFold(edition, "Enterprise")

	nodes := d.GetInt64OrDefault("number_of_nodes", 1)

	nodeType := d.Get("node_size").String()
	instanceType := strings.ReplaceAll(nodeType, "Standard_", "")
	instanceType = strings.ReplaceAll(instanceType, "_", " ")

	r := &azure.DataFactoryIntegrationRuntimeManaged{
		Address:         d.Address,
		Region:          region,
		Enterprise:      enterprise,
		LicenseIncluded: licenseIncluded,
		Instances:       nodes,
		InstanceType:    instanceType,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
