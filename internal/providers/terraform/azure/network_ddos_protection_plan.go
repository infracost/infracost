package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getNetworkDdosProtectionPlanRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_network_ddos_protection_plan",
		CoreRFunc: newNetworkDdosProtectionPlan,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newNetworkDdosProtectionPlan(d *schema.ResourceData) schema.CoreResource {
	region := d.Region
	return &azure.NetworkDdosProtectionPlan{
		Address: d.Address,
		Region:  region,
	}
}
