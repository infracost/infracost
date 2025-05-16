package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAzureRMMSSQLManagedInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_mssql_managed_instance",
		RFunc: newMSSQLManagedInstance,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newMSSQLManagedInstance(d *schema.ResourceData) schema.CoreResource {
	region := d.Region
	r := &azure.MSSQLManagedInstance{
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
