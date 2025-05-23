package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getSQLManagedInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_sql_managed_instance",
		RFunc: newSQLManagedInstance,
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

	r.SKU = d.Get("skuName").String()
	r.Cores = d.Get("vcores").Int()
	r.LicenseType = d.Get("licenseType").String()
	r.StorageAccountType = d.Get("storageAccountType").String()
	if r.StorageAccountType == "" {
		r.StorageAccountType = "LRS"
	}
	if r.StorageAccountType == "GRS" {
		r.StorageAccountType = "RA-GRS"
	}
	r.StorageSizeInGb = d.Get("storageSizeInGb").Int()

	return r
}
