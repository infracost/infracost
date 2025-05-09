package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func getStorageManagementPolicyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_storage_management_policy",
		RFunc: newStorageManagementPolicy,
		ReferenceAttributes: []string{
			"storage_account_id",
		},
	}
}

func newStorageManagementPolicy(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:      d.Address,
		NoPrice:   true,
		IsSkipped: true,
	}
}
