package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetNATGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_nat_gateway",
		RFunc: NewNATGateway,
	}
}

func NewNATGateway(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	args := &aws.NATGatewayArguments{
		Address: d.Address,
		Region:  region,
	}
	args.PopulateUsage(u)

	return aws.NewNATGateway(args)
}
