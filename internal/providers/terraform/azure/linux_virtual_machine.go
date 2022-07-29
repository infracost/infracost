package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAzureLinuxVirtualMachineRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_linux_virtual_machine",
		RFunc: NewAzureLinuxVirtualMachine,
		Notes: []string{
			"Non-standard images such as RHEL are not supported.",
			"Low priority, Spot and Reserved instances are not supported.",
		},
	}
}
func NewAzureLinuxVirtualMachine(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.LinuxVirtualMachine{
		Address:         d.Address,
		Region:          lookupRegion(d, []string{}),
		Size:            d.Get("size").String(),
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
	return r.BuildResource()
}
