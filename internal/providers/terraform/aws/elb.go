package aws

import "github.com/infracost/infracost/pkg/schema"

func NewELB(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	productFamily := "Load Balancer"
	costComponentName := "Per Classic Load Balancer"
	return newLBResource(d, u, productFamily, costComponentName)
}
