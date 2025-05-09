package aws

import (
	"github.com/infracost/infracost/internal/schema"
)

func getPipesPipeRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_pipes_pipe",
		ReferenceAttributes: []string{"targetParameters.0.ecsTaskParameters.0.taskDefinitionArn"},
		NoPrice:             true,
		Notes:               []string{"Free resource."},
	}
}
