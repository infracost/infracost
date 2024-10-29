package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getLogicAppIntegrationAccountRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_logic_app_integration_account",
		CoreRFunc: newLogicAppIntegrationAccount,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newLogicAppIntegrationAccount(d *schema.ResourceData) schema.CoreResource {
	region := d.Region

	return azure.NewLogicAppIntegrationAccount(d.Address, region, d.GetStringOrDefault("sku_name", "free"))
}
