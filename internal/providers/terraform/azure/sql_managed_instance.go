package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getSQLManagedInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_sql_managed_instance",
		CoreRFunc: newSQLManagedInstance,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newSQLManagedInstance(d *schema.ResourceData) schema.CoreResource {
	region := d.Region
	r := &azure.SQLManagedInstance{
		Address: d.Address,
		Region:  region,
	}

	r.SKU = d.Get("sku_name").String()
	r.Cores = d.Get("vcores").Int()
	r.LicenseType = d.Get("license_type").String()
	r.StorageAccountType = d.Get("storage_account_type").String()
	if r.StorageAccountType == "" {
		r.StorageAccountType = "LRS"
	}
	if r.StorageAccountType == "GRS" {
		r.StorageAccountType = "RA-GRS"
	}
	r.StorageSizeInGb = d.Get("storage_size_in_gb").Int()

	return r
}
