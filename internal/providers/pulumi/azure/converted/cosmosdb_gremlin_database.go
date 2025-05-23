package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMCosmosdbGremlinDatabaseRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_cosmosdb_gremlin_database",
		RFunc: NewAzureRMCosmosdb,
		ReferenceAttributes: []string{
			"account_name",
			"resource_group_name",
		},
	}
}
