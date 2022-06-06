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
		Name:                "aws:ec2/eip:Eip",
		ReferenceAttributes: eipReferences,
		RFunc:               NewEIP,
	}
}
func NewEIP(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var allocated bool

	if !d.IsEmpty("customerOwnedIpv4Pool") || !d.IsEmpty("instance") || !d.IsEmpty("networkInterface") {
		allocated = true
	}
	var region = d.Get("config.aws:region")
	r := &aws.EIP{
		Address:   d.Address,
		Region:    region.String(),
		Allocated: allocated,
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
