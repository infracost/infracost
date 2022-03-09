package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getComputeExternalVPNGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_external_vpn_gateway",
		RFunc: NewComputeExternalVPNGateway,
	}
}
func NewComputeExternalVPNGateway(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.ComputeExternalVPNGateway{Address: d.Address, Region: d.Get("region").String()}
	r.PopulateUsage(u)
	return r.BuildResource()
}
