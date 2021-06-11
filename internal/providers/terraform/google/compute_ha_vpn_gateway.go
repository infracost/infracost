package google

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetComputeHAVPNGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_ha_vpn_gateway",
		RFunc: NewComputeHAVPNGateway,
	}
}

func NewComputeHAVPNGateway(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	return &schema.Resource{
		Name: d.Address,
		SubResources: []*schema.Resource{
			networkEgress(region, u, "Network egress", "IPSec traffic", ComputeVPNGateway),
		},
	}
}
