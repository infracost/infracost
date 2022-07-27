package aws

import (
	"github.com/infracost/infracost/internal/schema"
)

// This is a free resource but needs it's own custom registry item to specify the custom ID lookup function.
func getECSTaskDefinitionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:    "aws_ecs_task_definition",
		NoPrice: true,
		Notes:   []string{"Free resource."},
		CustomRefIDFunc: func(d *schema.ResourceData) []string {
			refs := []string{d.Get("arn").String()}

			family := d.Get("family").String()
			if family != "" {
				refs = append(refs, family)
			}

			return refs
		},
	}
}
