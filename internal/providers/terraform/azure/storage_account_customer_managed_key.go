package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func getStorageAccountCustomerManagedKeyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name: "azurerm_storage_account_customer_managed_key",
		CoreRFunc: func(d *schema.ResourceData) schema.CoreResource {
			return schema.BlankCoreResource{
				Name: d.Address,
				Type: d.Type,
			}
		},
		ReferenceAttributes: []string{
			"storage_account_id",
		},
	}
}
