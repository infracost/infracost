package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getVirtualMachineRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_virtual_machine",
		RFunc: NewVirtualMachine,
	}
}
func NewVirtualMachine(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.VirtualMachine{
		Address:                    d.Address,
		Region:                     d.Region,
		StorageImageReferenceOffer: d.Get("storageImageReference.0.offer").String(),
		StorageOSDiskOSType:        d.Get("storageOsDisk.0.osType").String(),
		LicenseType:                d.Get("licenseType").String(),
		VMSize:                     d.Get("vmSize").String(),
		StoragesDiskData:           make([]*azure.ManagedDiskData, 0),
	}

	if len(d.Get("storageOsDisk").Array()) > 0 {
		storageData := d.Get("storageOsDisk").Array()[0]
		r.StorageOSDiskData = &azure.ManagedDiskData{
			DiskType:   storageData.Get("managed_disk_type").String(),
			DiskSizeGB: storageData.Get("disk_size_gb").Int(),
		}
	}

	if len(d.Get("storageDataDisk").Array()) > 0 {
		for _, s := range d.Get("storageDataDisk").Array() {
			r.StoragesDiskData = append(r.StoragesDiskData, &azure.ManagedDiskData{
				DiskType:   s.Get("managed_disk_type").String(),
				DiskSizeGB: s.Get("disk_size_gb").Int(),
			})
		}
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
