package aws

import (
	"github.com/infracost/infracost/internal/schema"
)

func getSubnetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_subnet",
		RFunc:               NewSubnet,
		ReferenceAttributes: []string{"aws_nat_gateway.subnet_id"},
	}
}

func NewSubnet(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:         d.Address,
		ResourceType: d.Type,
		Tags:         d.Tags,
		DefaultTags:  d.DefaultTags,
		IsSkipped:    true,
		NoPrice:      true,
		SkipMessage:  "Free resource.",
	}
}
