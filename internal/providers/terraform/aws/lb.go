package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getLBRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name: "aws_lb",
		ReferenceAttributes: []string{
			"subnet_mapping.#.allocation_id",
		},
		CoreRFunc: NewLB,
	}
}

func getALBRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_alb",
		CoreRFunc: NewLB,
	}
}

func NewLB(d *schema.ResourceData) schema.CoreResource {
	loadBalancerType := d.Get("load_balancer_type").String()
	if loadBalancerType == "" {
		// set the default load balancer type as given https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lb
		// this is done as parsing the raw HCL will not pick up the default but return a blank string.
		loadBalancerType = "application"
	}

	r := &aws.LB{
		Address:          d.Address,
		Region:           d.Get("region").String(),
		LoadBalancerType: loadBalancerType,
	}
	return r
}
