package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAutomationAccountRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_automation_account",
		RFunc: NewAutomationAccount,
	}
}
func NewAutomationAccount(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.AutomationAccount{Address: d.Address, Region: lookupRegion(d, []string{})}
	r.PopulateUsage(u)
	return r.BuildResource()
}
