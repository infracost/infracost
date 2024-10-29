package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAutomationAccountRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_automation_account",
		CoreRFunc: NewAutomationAccount,
	}
}
func NewAutomationAccount(d *schema.ResourceData) schema.CoreResource {
	r := &azure.AutomationAccount{Address: d.Address, Region: d.Region}
	return r
}
