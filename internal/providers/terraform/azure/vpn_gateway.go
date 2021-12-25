package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAzureRMVPNGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_vpn_gateway",
		RFunc: newVPNGateway,
	}
}

func newVPNGateway(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	v := &azure.VPNGateway{
		Address:    d.Address,
		Region:     d.Get("region").String(),
		ScaleUnits: d.Get("scale_unit").Int(),
		Type:       "S2S",
	}
	v.PopulateUsage(u)

	return v.BuildResource()
}
