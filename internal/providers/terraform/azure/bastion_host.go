package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getBastionHostRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_bastion_host",
		RFunc: NewBastionHost,
	}
}
func NewBastionHost(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.BastionHost{Address: d.Address, Region: lookupRegion(d, []string{})}
	r.PopulateUsage(u)
	return r.BuildResource()
}
