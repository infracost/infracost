package google

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetComputeExternalVPNGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_external_vpn_gateway",
		RFunc: NewComputeExternalVPNGateway,
	}
}

func NewComputeExternalVPNGateway(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	return &schema.Resource{
		Name: d.Address,
		SubResources: []*schema.Resource{
			networkEgress(region, u, "Network egress", "IPSec traffic", ComputeExternalVPNGateway),
		},
	}
}
