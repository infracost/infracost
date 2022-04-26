package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAutomationDSCNodeConfigurationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_automation_dsc_nodeconfiguration",
		RFunc: NewAutomationDSCNodeConfiguration,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}
func NewAutomationDSCNodeConfiguration(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.AutomationDSCNodeConfiguration{Address: d.Address, Region: lookupRegion(d, []string{"resource_group_name"})}
	r.PopulateUsage(u)
	return r.BuildResource()
}
