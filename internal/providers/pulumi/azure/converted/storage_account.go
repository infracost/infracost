package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getStorageAccountRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_storage_account",
		RFunc: newAzureRMStorageAccount,
		ReferenceAttributes: []string{
			"azurerm_storage_management_policy.storage_account_id",
		},
		CustomRefIDFunc: func(d *schema.ResourceData) []string {
			return []string{d.Get("name").String()}
		},
	}
}

func newAzureRMStorageAccount(d *schema.ResourceData) schema.CoreResource {
	region := d.Region

	accountKind := "StorageV2"
	if !d.IsEmpty("account_kind") {
		accountKind = d.Get("accountKind").String()
	}

	accountReplicationType := d.Get("accountReplicationType").String()
	switch strings.ToLower(accountReplicationType) {
	case "ragrs":
		accountReplicationType = "RA-GRS"
	case "ragzrs":
		accountReplicationType = "RA-GZRS"
	}

	accountTier := d.Get("accountTier").String()

	accessTier := "Hot"
	if !d.IsEmpty("access_tier") {
		accessTier = d.Get("accessTier").String()
	}

	nfsv3 := false
	if !d.IsEmpty("nfsv3_enabled") {
		nfsv3 = d.Get("nfsv3Enabled").Bool()
	}

	return &azure.StorageAccount{
		Address:                d.Address,
		Region:                 region,
		AccessTier:             accessTier,
		AccountKind:            accountKind,
		AccountReplicationType: accountReplicationType,
		AccountTier:            accountTier,
		NFSv3:                  nfsv3,
	}
}
