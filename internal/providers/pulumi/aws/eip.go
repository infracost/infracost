package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

var (
	eipReferences = []string{
		"awsNatGateway.allocation_id",
		"awsEipAssociation.allocation_id",
		"awsLb.subnetMapping.#.allocationId",
	}
)

func getEIPRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_ec2_eip",
		ReferenceAttributes: eipReferences,
		RFunc:               NewEIP,
	}
}
func NewEIP(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var allocated bool

	if !d.IsEmpty("customerOwnedIpv4Pool") || !d.IsEmpty("instance") || !d.IsEmpty("networkInterface") {
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
