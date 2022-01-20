package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getRoute53ZoneRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_route53_zone",
		RFunc: NewRoute53Zone,
	}
}

func NewRoute53Zone(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.Route53Zone{
		Address: d.Address,
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
