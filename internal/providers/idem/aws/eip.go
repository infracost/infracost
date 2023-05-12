package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

var (
	eipReferences = []string{
		"states.aws.ec2.nat_gateway.present:allocation_id",
		"states.aws.elbv2.load_balancer.present:subnet_mappings.#.allocation_id",
	}
)

func GetEIPRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "states.aws.ec2.elastic_ip.present",
		ReferenceAttributes: eipReferences,
		RFunc:               NewEIP,
	}
}
func NewEIP(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var allocated bool
	if len(d.References(eipReferences...)) > 0 {
		allocated = true
	}

	if !d.IsEmpty("customer_owned_ipv4_pool") || !d.IsEmpty("instance") || !d.IsEmpty("network_interface") {
		allocated = true
	}

	r := &aws.EIP{
		Address:   d.Address,
		Region:    d.Get("region").String(),
		Allocated: allocated,
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
