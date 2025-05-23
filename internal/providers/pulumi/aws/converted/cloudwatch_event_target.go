package aws

import (
	"github.com/infracost/infracost/internal/schema"
)

func getCloudwatchEventTargetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_cloudwatch_event_target",
		ReferenceAttributes: []string{"ecsTarget.0.taskDefinitionArn"},
		NoPrice:             true,
		Notes:               []string{"Free resource."},
	}
}
