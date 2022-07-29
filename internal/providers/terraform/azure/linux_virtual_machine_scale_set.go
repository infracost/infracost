package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAzureLinuxVirtualMachineScaleSetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_linux_virtual_machine_scale_set",
		RFunc: NewLinuxVirtualMachineScaleSet,
	}
}
func NewLinuxVirtualMachineScaleSet(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.LinuxVirtualMachineScaleSet{
		Address:         d.Address,
		Region:          lookupRegion(d, []string{}),
		SKU:             d.Get("sku").String(),
		UltraSSDEnabled: d.Get("additional_capabilities.0.ultra_ssd_enabled").Bool(),
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

	r.PopulateUsage(u)
	if u.IsEmpty("instances") {
		r.Instances = intPtr(d.Get("instances").Int())
	}
	return r.BuildResource()
}
