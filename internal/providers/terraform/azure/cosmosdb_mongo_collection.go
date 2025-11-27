package azure

import (
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
)

// GetAzureRMCosmosdbMongoCollectionRegistryItem returns the schema.RegistryItem for Azure Cosmos DB MongoDB collections.
func GetAzureRMCosmosdbMongoCollectionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_cosmosdb_mongo_collection",
		RFunc: NewAzureRMCosmosdbMongoCollection,
		ReferenceAttributes: []string{
			"account_name",
			"database_name",
			"resource_group_name",
		},
	}
}

// NewAzureRMCosmosdbMongoCollection processes and returns the schema.Resource for Azure Cosmos DB MongoDB collections.
func NewAzureRMCosmosdbMongoCollection(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	databaseRefs := d.References("database_name")
	if len(databaseRefs) == 0 {
		logging.Logger.Warn().Msgf("Skipping resource %s as its 'database_name' property could not be found.", d.Address)
		return nil
	}

	mongoDB := databaseRefs[0]
	accountRefs := mongoDB.References("account_name")
	if len(accountRefs) == 0 {
		logging.Logger.Warn().Msgf("Skipping resource %s as its 'database_name.account_name' property could not be found.", d.Address)
		return nil
	}

	account := accountRefs[0]
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: cosmosDBCostComponents(d, u, account),
	}
}
