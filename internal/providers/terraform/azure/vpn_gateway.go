package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getVPNGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_vpn_gateway",
		CoreRFunc: newVPNGateway,
	}
}

func newVPNGateway(d *schema.ResourceData) schema.CoreResource {
	v := &azure.VPNGateway{
		Address:    d.Address,
		Region:     d.Get("region").String(),
		ScaleUnits: d.GetInt64OrDefault("scale_unit", 1),
		Type:       "S2S",
	}

	return v
}
