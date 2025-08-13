package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getManagedDiskRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "Microsoft.Compute/disks",
		CoreRFunc: NewManagedDisk,
	}
}

func NewManagedDisk(d *schema.ResourceData) schema.CoreResource {
	r := &azure.ManagedDisk{
		Address: d.Address,
		Region:  d.Region,
		ManagedDiskData: azure.ManagedDiskData{
			DiskType:          d.Get("sku.name").String(),
			DiskSizeGB:        d.Get("properties.diskSizeGB").Int(),
			DiskIOPSReadWrite: d.Get("properties.diskIOPSReadWrite").Int(),
			DiskMBPSReadWrite: d.Get("properties.diskMBpsReadWrite").Int(),
		},
	}

	return r
}
