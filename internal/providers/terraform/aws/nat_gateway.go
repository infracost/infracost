package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getNATGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name: "aws_nat_gateway",
		ReferenceAttributes: []string{
			"allocation_id",
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
