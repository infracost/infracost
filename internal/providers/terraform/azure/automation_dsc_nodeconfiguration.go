package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMAutomationDscNodeconfigurationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_automation_dsc_nodeconfiguration",
		RFunc: NewAzureRMAutomationDscConfiguration,
	}
}
