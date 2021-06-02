package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMCosmosdbGremlinDatabaseRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_cosmosdb_gremlin_database",
		RFunc: NewAzureRMCosmosdbGremlinDatabase,
		ReferenceAttributes: []string{
			"account_name",
		},
	}
}

func NewAzureRMCosmosdbGremlinDatabase(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	account := d.References("account_name")[0]

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: cosmosDBCostComponents(d, u, account),
	}
}
