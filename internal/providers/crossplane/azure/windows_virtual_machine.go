package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getWindowsVirtualMachineRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "compute.azure.upbound.io/WindowsVirtualMachine",
		CoreRFunc: NewWindowsVirtualMachine,
		Notes: []string{
			"Low priority, Spot and Reserved instances are not supported.",
		},
	}
}

func NewWindowsVirtualMachine(d *schema.ResourceData) schema.CoreResource {
	forProvider := d.Get("forProvider")

	region := lookupRegion(d, []string{})
	
	r := &azure.WindowsVirtualMachine{
		Address:                               d.Address,
		Region:                                region,
		Size:                                  strings.ToLower(forProvider.Get("size").String()),
		LicenseType:                           forProvider.Get("licenseType").String(),
		AdditionalCapabilitiesUltraSSDEnabled: forProvider.Get("additionalCapabilities.0.ultraSsdEnabled").Bool(),
	}
	if len(forProvider.Get("osDisk").Array()) > 0 {
		diskData := forProvider.Get("osDisk").Array()[0]
		r.OSDiskData = &azure.ManagedDiskData{
			DiskType:   diskData.Get("storageAccountType").String(),
			DiskSizeGB: diskData.Get("diskSizeGb").Int(),
		}
	}
	return r
}
