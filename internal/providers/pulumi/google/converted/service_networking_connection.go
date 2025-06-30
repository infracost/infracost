package google

import (
	"github.com/infracost/infracost/internal/schema"
)

func getServiceNetworkingConnectionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_service_networking_connection",
		RFunc: newServiceNetworkingConnection,
	}
}

func newServiceNetworkingConnection(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	return &schema.Resource{
		Name: d.Address,
		SubResources: []*schema.Resource{
			networkEgress(region, u, "Network egress", "Traffic", ComputeVPNGateway),
		},
	}
}
