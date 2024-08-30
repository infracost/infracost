package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getLinuxVirtualMachineRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "compute.azure.upbound.io/LinuxVirtualMachine",
		CoreRFunc: NewAzureLinuxVirtualMachine,
		Notes: []string{
			"Non-standard images such as RHEL are not supported.",
			"Low priority, Spot and Reserved instances are not supported.",
		},
	}
}
func NewAzureLinuxVirtualMachine(d *schema.ResourceData) schema.CoreResource {
	forProvider := d.Get("forProvider")

	region := lookupRegion(d, []string{})

	r := &azure.LinuxVirtualMachine{
		Address:         d.Address,
		Region:          region,
		Size:            strings.ToLower(forProvider.Get("size").String()),
		UltraSSDEnabled: forProvider.Get("additionalCapabilities.0.ultraSsdEnabled").Bool(),
	}

	if len(d.Get("osDisk").Array()) > 0 {
		storageData := forProvider.Get("osDisk").Array()[0]
		r.OSDiskData = &azure.ManagedDiskData{
			DiskType:   storageData.Get("storageAccountType").String(),
			DiskSizeGB: storageData.Get("diskSizeGb").Int(),
		}
	}

	return r
}
