package aws

import (
	"github.com/infracost/infracost/internal/schema"
)

func getECSClusterCapacityProvidersRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_ecs_cluster_capacity_providers",
		RFunc:               NewECSClusterCapacityProviders,
		ReferenceAttributes: []string{"cluster_name"},
	}
}

func NewECSClusterCapacityProviders(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
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
