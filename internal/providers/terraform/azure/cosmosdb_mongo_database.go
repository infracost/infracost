package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMCosmosdbMongoDatabaseRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_cosmosdb_mongo_database",
		RFunc: NewAzureRMCosmosdb,
		ReferenceAttributes: []string{
			"account_name",
			"resource_group_name",
		},
	}
}
