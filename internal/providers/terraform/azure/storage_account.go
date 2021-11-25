package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAzureRMStorageAccountRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_storage_account",
		RFunc: newAzureRMStorageAccount,
	}
}

func newAzureRMStorageAccount(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})

	accountKind := "StorageV2"
	if !d.IsEmpty("account_kind") {
		accountKind = d.Get("account_kind").String()
	}

	accountReplicationType := d.Get("account_replication_type").String()
	switch strings.ToLower(accountReplicationType) {
	case "ragrs":
		accountReplicationType = "RA-GRS"
	case "ragzrs":
		accountReplicationType = "RA-GZRS"
	}

	accountTier := d.Get("account_tier").String()

	accessTier := "Hot"
	if !d.IsEmpty("access_tier") {
		accessTier = d.Get("access_tier").String()
	}

	r := &azure.StorageAccount{
		Address:                d.Address,
		Region:                 region,
		AccessTier:             accessTier,
		AccountKind:            accountKind,
		AccountReplicationType: accountReplicationType,
		AccountTier:            accountTier,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
