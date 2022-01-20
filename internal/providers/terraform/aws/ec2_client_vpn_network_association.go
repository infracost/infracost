package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getEC2ClientVPNNetworkAssociationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ec2_client_vpn_network_association",
		RFunc: NewEC2ClientVPNNetworkAssociation,
	}
}
func NewEC2ClientVPNNetworkAssociation(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.EC2ClientVPNNetworkAssociation{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
