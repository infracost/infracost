package aws

import "infracost/pkg/schema"

var ResourceRegistry map[string]schema.ResourceFunc = map[string]schema.ResourceFunc{
	"aws_autoscaling_group": NewAutoscalingGroup,
	"aws_dynamodb_table":    NewDynamoDBTable,
	"aws_ecs_service":       NewEcsService,
	"aws_instance":          NewInstance,
	"aws_nat_gateway":       NewNatGateway,
}
