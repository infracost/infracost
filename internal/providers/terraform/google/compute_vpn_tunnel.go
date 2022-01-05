package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getComputeVPNTunnelRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_vpn_tunnel",
		RFunc: NewComputeVpnTunnel,
	}
}
func NewComputeVpnTunnel(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.ComputeVpnTunnel{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
