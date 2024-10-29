package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getBastionHostRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_bastion_host",
		CoreRFunc: NewBastionHost,
	}
}
func NewBastionHost(d *schema.ResourceData) schema.CoreResource {
	r := &azure.BastionHost{Address: d.Address, Region: d.Region}
	return r
}
