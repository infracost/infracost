package aws

import (
	"github.com/infracost/infracost/internal/schema"
)

func getSchedulerScheduleRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_scheduler_schedule",
		ReferenceAttributes: []string{"ecsParameters.0.taskDefinitionArn"},
		NoPrice:             true,
		Notes:               []string{"Free resource."},
	}
}
