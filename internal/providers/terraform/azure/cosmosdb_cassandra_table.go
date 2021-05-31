package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMCosmosdbCassandraTableRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_cosmosdb_cassandra_table",
		RFunc: NewAzureCosmosdbCassandraTable,
		ReferenceAttributes: []string{
			"account_name",
			"cassandra_keyspace_id",
		},
	}
}

func NewAzureCosmosdbCassandraTable(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	keyspace := d.References("cassandra_keyspace_id")[0]
	account := keyspace.References("account_name")[0]

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: cosmosDBCostComponents(d, u, account),
	}
}
