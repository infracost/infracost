package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getEC2ClientVPNEndpointRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ec2_client_vpn_endpoint",
		RFunc: NewEc2ClientVpnEndpoint,
	}
}
func NewEc2ClientVpnEndpoint(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.EC2ClientVPNEndpoint{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
