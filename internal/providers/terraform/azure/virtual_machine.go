package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getVirtualMachineRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_virtual_machine",
		RFunc: NewVirtualMachine,
	}
}
func NewVirtualMachine(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.VirtualMachine{
		Address:                    d.Address,
		Region:                     lookupRegion(d, []string{}),
		StorageImageReferenceOffer: d.Get("storage_image_reference.0.offer").String(),
		StorageOSDiskOSType:        d.Get("storage_os_disk.0.os_type").String(),
		LicenseType:                d.Get("license_type").String(),
		VMSize:                     d.Get("vm_size").String(),
		StoragesDiskData:           make([]*azure.ManagedDiskData, 0),
	}

	if len(d.Get("storage_os_disk").Array()) > 0 {
		storageData := d.Get("storage_os_disk").Array()[0]
		r.StorageOSDiskData = &azure.ManagedDiskData{
			DiskType:          storageData.Get("managed_disk_type").String(),
			DiskSizeGB:        storageData.Get("disk_size_gb").Int(),
			DiskIOPSReadWrite: storageData.Get("disk_iops_read_write").Int(),
			DiskMBPSReadWrite: storageData.Get("disk_mbps_read_write").Int(),
		}
	}

	if len(d.Get("os_disk").Array()) > 0 {
		storageData := d.Get("os_disk").Array()[0]
		r.OSDiskData = &azure.ManagedDiskData{
			DiskType:          storageData.Get("managed_disk_type").String(),
			DiskSizeGB:        storageData.Get("disk_size_gb").Int(),
			DiskIOPSReadWrite: storageData.Get("disk_iops_read_write").Int(),
			DiskMBPSReadWrite: storageData.Get("disk_mbps_read_write").Int(),
		}
	}

	if len(d.Get("storage_data_disk").Array()) > 0 {
		for _, s := range d.Get("storage_data_disk").Array() {
			r.StoragesDiskData = append(r.StoragesDiskData, &azure.ManagedDiskData{
				DiskType:          s.Get("managed_disk_type").String(),
				DiskSizeGB:        s.Get("disk_size_gb").Int(),
				DiskIOPSReadWrite: s.Get("disk_iops_read_write").Int(),
				DiskMBPSReadWrite: s.Get("disk_mbps_read_write").Int(),
			})
		}
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
