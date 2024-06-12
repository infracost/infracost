package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAutomationDSCConfigurationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_automation_dsc_configuration",
		CoreRFunc: NewAutomationDSCConfiguration,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}
func NewAutomationDSCConfiguration(d *schema.ResourceData) schema.CoreResource {
	r := &azure.AutomationDSCConfiguration{Address: d.Address, Region: d.Region}
	return r
}
