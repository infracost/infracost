package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getComputeVPNTunnelRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_compute_vpn_tunnel",
		CoreRFunc: NewComputeVPNTunnel,
	}
}

func NewComputeVPNTunnel(d *schema.ResourceData) schema.CoreResource {
	r := &google.ComputeVPNTunnel{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	return r
}
