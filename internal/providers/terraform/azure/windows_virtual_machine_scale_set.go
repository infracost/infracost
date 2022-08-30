package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func getWindowsVirtualMachineScaleSetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_windows_virtual_machine_scale_set",
		RFunc: NewWindowsVirtualMachineScaleSet,
	}
}
func NewWindowsVirtualMachineScaleSet(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.WindowsVirtualMachineScaleSet{
		Address:                               d.Address,
		Region:                                lookupRegion(d, []string{}),
		SKU:                                   d.Get("sku").String(),
		LicenseType:                           d.Get("license_type").String(),
		AdditionalCapabilitiesUltraSSDEnabled: d.Get("additional_capabilities.0.ultra_ssd_enabled").Bool(),
	}
	if len(d.Get("os_disk").Array()) > 0 {
		diskData := d.Get("os_disk").Array()[0]
		r.OSDiskData = &azure.ManagedDiskData{
			DiskType:          diskData.Get("managed_disk_type").String(),
			DiskSizeGB:        diskData.Get("disk_size_gb").Int(),
			DiskIOPSReadWrite: diskData.Get("disk_iops_read_write").Int(),
			DiskMBPSReadWrite: diskData.Get("disk_mbps_read_write").Int(),
		}
	}
	r.PopulateUsage(u)
	if u.Get("instances").Type == gjson.Null || u.Get("instances").Int() == 0 {
		r.Instances = intPtr(d.Get("instances").Int())
	}
	return r.BuildResource()
}
