package azure

import (
	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"
)

func GetAzureRMCosmosdbCassandraTableRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_cosmosdb_cassandra_table",
		RFunc: NewAzureRMCosmosdbCassandraTable,
		ReferenceAttributes: []string{
			"account_name",
			"cassandra_keyspace_id",
		},
	}
}

func NewAzureRMCosmosdbCassandraTable(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	if len(d.References("cassandra_keyspace_id")) > 0 {
		keyspace := d.References("cassandra_keyspace_id")[0]
		if len(keyspace.References("account_name")) > 0 {
			account := keyspace.References("account_name")[0]
			return &schema.Resource{
				Name:           d.Address,
				CostComponents: cosmosDBCostComponents(d, u, account),
			}
		}
		log.Warnf("Skipping resource %s as its 'cassandra_keyspace_id.account_name' property could not be found.", d.Address)
		return nil
	}
	log.Warnf("Skipping resource %s as its 'cassandra_keyspace_id' property could not be found.", d.Address)
	return nil
}
