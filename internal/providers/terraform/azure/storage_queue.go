package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getStorageQueueRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_storage_queue",
		CoreRFunc: newStorageQueue,
		ReferenceAttributes: []string{
			"storage_account_name",
		},
	}
}

func newStorageQueue(d *schema.ResourceData) schema.CoreResource {
	region := lookupRegion(d, []string{"storage_account_name"})

	accountReplicationType := "LRS"

	if len(d.References("storage_account_name")) > 0 {
		storageAccount := d.References("storage_account_name")[0]
		accountReplicationType = storageAccount.Get("account_replication_type").String()
	}

	switch strings.ToLower(accountReplicationType) {
	case "ragrs":
		accountReplicationType = "RA-GRS"
	case "ragzrs":
		accountReplicationType = "RA-GZRS"
	}

	return &azure.StorageQueue{
		Address:                d.Address,
		Region:                 region,
		AccountReplicationType: accountReplicationType,
	}
}
