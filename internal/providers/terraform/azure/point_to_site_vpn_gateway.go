package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAzureRMPointToSiteVpnGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_point_to_site_vpn_gateway",
		RFunc: newPointToSiteVpnGateway,
	}
}

func newPointToSiteVpnGateway(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	p := &azure.VPNGateway{
		Address:    d.Address,
		Region:     d.Get("region").String(),
		ScaleUnits: d.Get("scale_unit").Int(),
		Type:       "P2S",
	}
	p.PopulateUsage(u)

	return p.BuildResource()
}
