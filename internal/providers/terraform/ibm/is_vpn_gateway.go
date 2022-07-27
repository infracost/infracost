package ibm

import (
	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
)

func getIsVpnGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "ibm_is_vpn_gateway",
		RFunc: newIsVpnGateway,
	}
}

func newIsVpnGateway(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	r := &ibm.IsVpnGateway{
		Address: d.Address,
		Region:  region,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
