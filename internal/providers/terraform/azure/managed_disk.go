package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getManagedDiskRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_managed_disk",
		RFunc: NewManagedDisk,
	}
}

func NewManagedDisk(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.ManagedDisk{
		Address: d.Address,
		Region:  lookupRegion(d, []string{}),
		ManagedDiskData: azure.ManagedDiskData{
			DiskType:          d.Get("storage_account_type").String(),
			DiskSizeGB:        d.Get("disk_size_gb").Int(),
			DiskIOPSReadWrite: d.Get("disk_iops_read_write").Int(),
			DiskMBPSReadWrite: d.Get("disk_mbps_read_write").Int(),
		},
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
