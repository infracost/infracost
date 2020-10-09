package aws

import "github.com/infracost/infracost/internal/schema"

func GetELBRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_elb",
		RFunc: NewELB,
	}
}

func NewELB(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	productFamily := "Load Balancer"
	costComponentName := "Per Classic Load Balancer"
	return newLBResource(d, productFamily, costComponentName)
}
