package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getComputeVPNGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_vpn_gateway",
		RFunc: NewComputeVpnGateway,
	}
}
func NewComputeVpnGateway(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.ComputeVpnGateway{Address: d.Address, Region: d.Get("region").String()}
	r.PopulateUsage(u)
	return r.BuildResource()
}
