package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMCosmosdbMongoCollectionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_cosmosdb_mongo_collection",
		RFunc: NewAzureCosmosdbMongoCollection,
		ReferenceAttributes: []string{
			"account_name",
			"database_name",
		},
	}
}

func NewAzureCosmosdbMongoCollection(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	mongoDB := d.References("database_name")[0]
	account := mongoDB.References("account_name")[0]

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: cosmosDBCostComponents(d, u, account),
	}
}
