package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

// This is a free resource but needs it's own custom registry item to specify the custom ID lookup function.
func getCosmosDBAccountRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:    "azurerm_cosmosdb_account",
		NoPrice: true,
		Notes:   []string{"Free resource."},

		CustomRefIDFunc: func(d *schema.ResourceData) []string {
			return []string{d.Get("name").String()}
		},
	}
}
