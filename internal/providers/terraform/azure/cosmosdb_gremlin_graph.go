package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMCosmosdbGremlinGraphRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_cosmosdb_gremlin_graph",
		RFunc: NewAzureCosmosdbGremlinGraph,
		ReferenceAttributes: []string{
			"account_name",
		},
	}
}

func NewAzureCosmosdbGremlinGraph(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: cosmosDBCostComponents(d, u),
	}
}
