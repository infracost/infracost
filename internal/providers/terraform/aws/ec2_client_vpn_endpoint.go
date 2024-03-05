package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getEC2ClientVPNEndpointRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_ec2_client_vpn_endpoint",
		CoreRFunc: NewEc2ClientVpnEndpoint,
	}
}
func NewEc2ClientVpnEndpoint(d *schema.ResourceData) schema.CoreResource {
	r := &aws.EC2ClientVPNEndpoint{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}
	return r
}
