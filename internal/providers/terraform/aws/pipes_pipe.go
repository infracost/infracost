package aws

import (
	"github.com/infracost/infracost/internal/schema"
)

func getPipesPipeRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_pipes_pipe",
		ReferenceAttributes: []string{"target_parameters.0.ecs_task_parameters.0.task_definition_arn"},
		NoPrice:             true,
		Notes:               []string{"Free resource."},
	}
}
