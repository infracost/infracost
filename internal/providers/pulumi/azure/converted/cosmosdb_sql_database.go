package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMCosmosdbSQLDatabaseRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_cosmosdb_sql_database",
		RFunc: NewAzureRMCosmosdb,
		ReferenceAttributes: []string{
			"account_name",
			"resource_group_name",
		},
	}
}
