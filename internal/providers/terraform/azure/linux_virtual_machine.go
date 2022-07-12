package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAzureLinuxVirtualMachineRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_linux_virtual_machine",
		RFunc: NewLB,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewLinuxVirtualMachine(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.LinuxVirtualMachine{
		Address: d.Address,
		Region: lookupRegion(d, []string{"resource_group_name"}),
		Size: d.Get("size").String(),
		UltraSSDEnabled: d.Get("additional_capabilities.0.ultra_ssd_enabled").Bool(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
