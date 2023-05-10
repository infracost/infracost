package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetNATGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name: "states.aws.ec2.nat_gateway.present",
		ReferenceAttributes: []string{
			"states.aws.ec2.elastic_ip.present:allocation_id",
		},
		RFunc: NewNATGateway,
	}
}

func NewNATGateway(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	a := &aws.NATGateway{
		Address: d.Address,
		Region:  region,
	}
	a.PopulateUsage(u)

	return a.BuildResource()
}
