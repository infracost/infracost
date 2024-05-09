package azure

import (
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
)

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

func NewAzureRMCosmosdbMongoCollection(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	if len(d.References("database_name")) > 0 {
		mongoDB := d.References("database_name")[0]
		if len(mongoDB.References("account_name")) > 0 {
			account := mongoDB.References("account_name")[0]
			return &schema.Resource{
				Name:           d.Address,
				CostComponents: cosmosDBCostComponents(d, u, account),
			}
		}
		logging.Logger.Warn().Msgf("Skipping resource %s as its 'database_name.account_name' property could not be found.", d.Address)
		return nil
	}
	logging.Logger.Warn().Msgf("Skipping resource %s as its 'database_name' property could not be found.", d.Address)
	return nil
}
