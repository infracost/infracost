package aws

import (
	"github.com/infracost/infracost/internal/schema"
)

func getECSClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ecs_cluster",
		RFunc: NewECSCluster,
		// this is a reverse reference, it depends on the aws_ecs_cluster_capacity_provider RegistryItem
		// defining "cluster_name" as a ReferenceAttribute
		ReferenceAttributes: []string{"awsEcsClusterCapacityProviders.clusterName"},
		CustomRefIDFunc: func(d *schema.ResourceData) []string {
			name := d.Get("name").String()
			if name != "" {
				return []string{name}
			}

			return nil
		},
	}
}

func NewECSCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
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
