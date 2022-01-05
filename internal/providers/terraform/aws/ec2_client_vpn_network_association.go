package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getEC2ClientVPNNetworkAssociationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ec2_client_vpn_network_association",
		RFunc: NewEc2ClientVpnNetworkAssociation,
	}
}
func NewEc2ClientVpnNetworkAssociation(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.Ec2ClientVpnNetworkAssociation{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
