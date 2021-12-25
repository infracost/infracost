package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetEC2ClientVPNEndpointRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ec2_client_vpn_endpoint",
		RFunc: NewEC2ClientVPNEndpoint,
	}
}
func NewEC2ClientVPNEndpoint(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.EC2ClientVPNEndpoint{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
