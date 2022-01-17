package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getELBRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_elb",
		RFunc: NewElb,
	}
}
func NewElb(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.Elb{Address: strPtr(d.Address)}
	r.PopulateUsage(u)
	return r.BuildResource()
}
