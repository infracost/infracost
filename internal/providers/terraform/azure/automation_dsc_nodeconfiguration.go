package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMAutomationDscNodeconfigurationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_automation_dsc_nodeconfiguration",
		RFunc: NewAzureRMAutomationDscNodeconfiguration,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewAzureRMAutomationDscNodeconfiguration(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {

	group := d.References("resource_group_name")[0]

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: nodesCostComponent(d, u, group),
	}
}
