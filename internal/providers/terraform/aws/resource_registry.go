package aws

import "infracost/pkg/schema"

var ResourceRegistry map[string]schema.ResourceFunc = map[string]schema.ResourceFunc{
	"aws_instance":    AwsInstance,
	"aws_nat_gateway": AwsNatGateway,
	"aws_ecs_service": AwsEcsService,
}
