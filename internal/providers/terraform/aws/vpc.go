package aws

import (
	"github.com/infracost/infracost/internal/schema"
)

func getVPCRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_vpc",
		RFunc:               NewVPC,
		ReferenceAttributes: []string{"aws_vpc_endpoint.vpc_id"},
	}
}

func NewVPC(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
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
