package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getManagedDiskRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_managed_disk",
		RFunc: NewManagedDisk,
	}
}

func NewManagedDisk(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.ManagedDisk{
		Address: d.Address,
		Region:  d.Region,
		ManagedDiskData: azure.ManagedDiskData{
			DiskType:          d.Get("storageAccountType").String(),
			DiskSizeGB:        d.Get("diskSizeGb").Int(),
			DiskIOPSReadWrite: d.Get("diskIopsReadWrite").Int(),
			DiskMBPSReadWrite: d.Get("diskMbpsReadWrite").Int(),
		},
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
