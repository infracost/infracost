package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getELBRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_elb",
		RFunc: NewELB,
	}
}
func NewELB(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.ELB{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
