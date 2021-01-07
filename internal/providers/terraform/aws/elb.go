package aws

import "github.com/infracost/infracost/internal/schema"

func GetELBRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_elb",
		RFunc: NewELB,
	}
}

func NewELB(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	productFamily := "Load Balancer"
	costComponentName := "Classic load balancer"
	return newLBResource(d, productFamily, costComponentName)
}
