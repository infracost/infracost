package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getVPNConnectionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_vpn_connection",
		CoreRFunc: NewVPNConnection,
	}
}
func NewVPNConnection(d *schema.ResourceData) schema.CoreResource {
	r := &aws.VPNConnection{Address: d.Address, TransitGatewayID: d.Get("transit_gateway_id").String(), Region: d.Get("region").String()}
	return r
}
