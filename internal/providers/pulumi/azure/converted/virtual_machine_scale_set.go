package azure

import (
	"strings"

	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getVirtualMachineScaleSetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_virtual_machine_scale_set",
		RFunc: NewVirtualMachineScaleSet,
	}
}
func NewVirtualMachineScaleSet(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.VirtualMachineScaleSet{
		Address:     d.Address,
		Region:      d.Region,
		SKUName:     d.Get("sku.0.name").String(),
		SKUCapacity: d.Get("sku.0.capacity").Int(),
		LicenseType: d.Get("licenseType").String(),
		IsDevTest:   d.ProjectMetadata["isProduction"] == "false",
	}

	if !d.IsEmpty("os_profile_windows_config") {
		r.IsWindows = true
	}
	if d.Get("storageProfileOsDisk.0.osType").Type != gjson.Null {
		if strings.ToLower(d.Get("storageProfileOsDisk.0.osType").String()) == "windows" {
			r.IsWindows = true
		}
	}
	if d.Get("storageProfileImageReference.0.offer").Type != gjson.Null {
		if strings.ToLower(d.Get("storageProfileImageReference.0.offer").String()) == "windowsserver" {
			r.IsWindows = true
		}
	}

	if len(d.Get("storageProfileOsDisk").Array()) > 0 {
		storageData := d.Get("storageProfileOsDisk").Array()[0]
		r.StorageProfileOSDiskData = &azure.ManagedDiskData{
			DiskType:   storageData.Get("managed_disk_type").String(),
			DiskSizeGB: storageData.Get("disk_size_gb").Int(),
		}
	}

	if len(d.Get("storageProfileDataDisk").Array()) > 0 {
		for _, s := range d.Get("storageProfileDataDisk").Array() {
			if s.Get("managed_disk_type").Type == gjson.Null {
				continue
			}
			r.StorageProfileOSDisksData = append(r.StorageProfileOSDisksData, &azure.ManagedDiskData{
				DiskType:   s.Get("managed_disk_type").String(),
				DiskSizeGB: s.Get("disk_size_gb").Int(),
			})
		}
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
