package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getManagedDiskRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_managed_disk",
		CoreRFunc: NewManagedDisk,
	}
}

func NewManagedDisk(d *schema.ResourceData) schema.CoreResource {
	r := &azure.ManagedDisk{
		Address: d.Address,
		Region:  d.Region,
		ManagedDiskData: azure.ManagedDiskData{
			DiskType:          d.Get("storage_account_type").String(),
			DiskSizeGB:        d.Get("disk_size_gb").Int(),
			DiskIOPSReadWrite: d.Get("disk_iops_read_write").Int(),
			DiskMBPSReadWrite: d.Get("disk_mbps_read_write").Int(),
		},
	}

	return r
}
