package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getVPNGatewayConnectionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_vpn_gateway_connection",
		CoreRFunc: newVPNGatewayConnection,
	}
}

func newVPNGatewayConnection(d *schema.ResourceData) schema.CoreResource {
	v := &azure.VPNGatewayConnection{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	return v
}
