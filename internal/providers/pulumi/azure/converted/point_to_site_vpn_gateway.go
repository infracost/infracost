package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getPointToSiteVpnGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_point_to_site_vpn_gateway",
		RFunc: newPointToSiteVpnGateway,
	}
}

func newPointToSiteVpnGateway(d *schema.ResourceData) schema.CoreResource {
	p := &azure.VPNGateway{
		Address:    d.Address,
		Region:     d.Get("region").String(),
		ScaleUnits: d.Get("scaleUnit").Int(),
		Type:       "P2S",
	}

	return p
}
