package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getStorageTableRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_storage_table",
		CoreRFunc: newStorageTable,
		ReferenceAttributes: []string{
			"storage_account_name",
		},
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			return lookupRegion(d, []string{"storage_account_name"})
		},
	}
}

func newStorageTable(d *schema.ResourceData) schema.CoreResource {
	region := d.Region

	accountReplicationType := "LRS"
	hasCustomerManagedKey := false

	if len(d.References("storage_account_name")) > 0 {
		storageAccount := d.References("storage_account_name")[0]
		accountReplicationType = storageAccount.Get("account_replication_type").String()
		hasCustomerManagedKey = storageAccount.Get("customer_managed_key").Exists()

		if storageAccount.References("azurerm_storage_account_customer_managed_key.storage_account_id") != nil {
			hasCustomerManagedKey = true
		}
	}

	switch strings.ToLower(accountReplicationType) {
	case "ragrs":
		accountReplicationType = "RA-GRS"
	case "ragzrs":
		accountReplicationType = "RA-GZRS"
	}

	return &azure.StorageTable{
		Address:                d.Address,
		Region:                 region,
		AccountReplicationType: accountReplicationType,
		HasCustomerManagedKey:  hasCustomerManagedKey,
	}
}