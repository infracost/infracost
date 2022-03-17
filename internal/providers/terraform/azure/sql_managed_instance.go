package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAzureRMSQLManagedInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_sql_managed_instance",
		RFunc: newSQLManagedInstance,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newSQLManagedInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{"resource_group_name"})
	r := &azure.SQLManagedInstance{
		Address: d.Address,
		Region:  region,
	}

	r.SKU = d.Get("sku_name").String()
	r.Cores = d.Get("vcores").Int()
	r.LicenceType = d.Get("license_type").String()
	r.StorageAccountType = d.Get("storage_account_type").String()
	if r.StorageAccountType == "" {
		r.StorageAccountType = "LRS"
	}
	if r.StorageAccountType == "GRS" {
		r.StorageAccountType = "RA-GRS"
	}
	r.StorageSizeInGb = d.Get("storage_size_in_gb").Int()

	r.PopulateUsage(u)

	return r.BuildResource()
}
