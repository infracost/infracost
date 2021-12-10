package aws

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetNATGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_nat_gateway",
		RFunc: NewNATGateway,
	}
}

func NewNATGateway(ctx *config.ProjectContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	a := &aws.NATGateway{
		Address: d.Address,
		Region:  region,
	}
	a.PopulateUsage(u)

	return a.BuildResource(ctx)
}
