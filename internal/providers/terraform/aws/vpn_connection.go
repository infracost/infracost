package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getVPNConnectionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_vpn_connection",
		RFunc: NewVpnConnection,
	}
}
func NewVpnConnection(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.VpnConnection{Address: d.Address, TransitGatewayID: d.Get("transit_gateway_id").String(), Region: d.Get("region").String()}
	r.PopulateUsage(u)
	return r.BuildResource()
}
