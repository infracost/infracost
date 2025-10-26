package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAutomationWatcherRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_automation_watcher",
		CoreRFunc: NewAutomationWatcher,
	}
}

func NewAutomationWatcher(d *schema.ResourceData) schema.CoreResource {
	r := &azure.AutomationWatcher{Address: d.Address, Region: d.Region}
	return r
}
