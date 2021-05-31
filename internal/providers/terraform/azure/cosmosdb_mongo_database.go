package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMCosmosdbMongoDatabaseRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_cosmosdb_mongo_database",
		RFunc: NewAzureCosmosdbMongoDatabase,
		ReferenceAttributes: []string{
			"account_name",
		},
	}
}

func NewAzureCosmosdbMongoDatabase(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	account := d.References("account_name")[0]

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: cosmosDBCostComponents(d, u, account),
	}
}
