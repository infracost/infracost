package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getLinuxVirtualMachineRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "Microsoft.Compute/virtualMachines/Linux",
		CoreRFunc: NewAzureLinuxVirtualMachine,
		Notes: []string{
			"Non-standard images such as RHEL are not supported.",
			"Low priority, Spot and Reserved instances are not supported.",
		},
	}
}
func NewAzureLinuxVirtualMachine(d *schema.ResourceData) schema.CoreResource {
	r := &azure.LinuxVirtualMachine{
		Address:         d.Address,
		Region:          d.Region,
		Size:            d.Get("properties.hardwareProfile.vmSize").String(),
		UltraSSDEnabled: d.Get("properties.additionalCapabilities.ultraSSDEnabled").Bool(),
	}

	if len(d.Get("properties.storageProfile.osDisk").Array()) > 0 {
		storageData := d.Get("properties.storageProfile.osDisk").Array()[0]
		r.OSDiskData = &azure.ManagedDiskData{
			DiskType:   storageData.Get("managedDisk.storageAccountType").String(),
			DiskSizeGB: storageData.Get("diskSizeGB").Int(),
		}
	}

	return r
}
