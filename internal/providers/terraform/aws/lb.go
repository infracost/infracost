package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getLBRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_lb",
		RFunc: NewLB,
	}
}

func GetALBRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_alb",
		RFunc: NewLB,
	}
}

func NewLB(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.LB{
		Address:          d.Address,
		Region:           d.Get("region").String(),
		LoadBalancerType: d.Get("load_balancer_type").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
