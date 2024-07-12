package aws

import (
	"github.com/infracost/infracost/internal/schema"
)

func getSchedulerScheduleRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_scheduler_schedule",
		ReferenceAttributes: []string{"ecs_parameters.0.task_definition_arn"},
		NoPrice:             true,
		Notes:               []string{"Free resource."},
	}
}
