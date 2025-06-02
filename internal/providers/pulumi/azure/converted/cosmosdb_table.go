package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMCosmosdbTableRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_cosmosdb_table",
		RFunc: NewAzureRMCosmosdb,
		ReferenceAttributes: []string{
			"account_name",
			"resource_group_name",
		},
	}
}
