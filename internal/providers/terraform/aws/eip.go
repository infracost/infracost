package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

var (
	eipReferences = []string{
		"aws_nat_gateway.allocation_id",
		"aws_eip_association.allocation_id",
		"aws_lb.subnet_mapping.#.allocation_id",
	}
)

func getEIPRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_eip",
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
