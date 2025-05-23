package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMCosmosdbGremlinGraphRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_cosmosdb_gremlin_graph",
		RFunc: NewAzureRMCosmosdb,
		ReferenceAttributes: []string{
			"account_name",
			"resource_group_name",
		},
	}
}
