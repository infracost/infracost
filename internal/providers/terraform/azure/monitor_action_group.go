package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getMonitorActionGroupRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_monitor_action_group",
		CoreRFunc: newMonitorActionGroup,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newMonitorActionGroup(d *schema.ResourceData) schema.CoreResource {
	region := lookupRegion(d, []string{"resource_group_name"})
	return &azure.MonitorActionGroup{
		Address: d.Address,
		Region:  region,
	}
}
