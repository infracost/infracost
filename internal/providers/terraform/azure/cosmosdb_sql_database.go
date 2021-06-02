package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMCosmosdbSQLDatabaseRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_cosmosdb_sql_database",
		RFunc: NewAzureRMCosmosdbSQLDatabase,
		ReferenceAttributes: []string{
			"account_name",
		},
	}
}

func NewAzureRMCosmosdbSQLDatabase(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	account := d.References("account_name")[0]

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: cosmosDBCostComponents(d, u, account),
	}
}
