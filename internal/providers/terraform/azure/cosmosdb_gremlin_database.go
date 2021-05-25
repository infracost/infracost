package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMCosmosdbGremlinDatabaseRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_cosmosdb_gremlin_database",
		RFunc: NewAzureCosmosdbGremlinDatabase,
		ReferenceAttributes: []string{
			"account_name",
		},
	}
}

func NewAzureCosmosdbGremlinDatabase(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: cosmosDBCostComponents(d, u),
	}
}
