package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAutomationDSCNodeConfigurationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_automation_dsc_nodeconfiguration",
		CoreRFunc: NewAutomationDSCNodeConfiguration,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}
func NewAutomationDSCNodeConfiguration(d *schema.ResourceData) schema.CoreResource {
	r := &azure.AutomationDSCNodeConfiguration{Address: d.Address, Region: d.Region}
	return r
}
