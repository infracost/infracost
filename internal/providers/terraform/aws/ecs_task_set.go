package aws

import (
	"github.com/infracost/infracost/internal/schema"
)

func getECSTaskSet() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name: "aws_ecs_task_set",
		RFunc: func(d *schema.ResourceData, _ *schema.UsageData) *schema.Resource {
			return &schema.Resource{
				Name:         d.Address,
				ResourceType: d.Type,
				Tags:         d.Tags,
				DefaultTags:  d.DefaultTags,
				IsSkipped:    true,
				NoPrice:      true,
				SkipMessage:  "Free resource.",
			}
		},
		ReferenceAttributes: []string{"service", "cluster", "task_definition"},
	}
}
