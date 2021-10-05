package aws_test

import (
	"testing"

	resources "github.com/infracost/infracost/internal/resources/aws"
)

// Tests LaunchConfiguration as a side effect.
func TestAutoscalingGroupOSWithLaunchConfiguration(t *testing.T) {
	stub := stubAWS(t)
	defer stub.Close()

	stubDescribeImages(stub, "ami-0227c65b90645ae0c", "RunInstances:0002")

	args := resources.AutoscalingGroup{
		LaunchConfiguration: &resources.LaunchConfiguration{AMI: "ami-0227c65b90645ae0c"},
	}
	resource := args.BuildResource()
	estimates := newEstimates(stub.ctx, t, resource)
	estimates.mustHave("operating_system", "windows")
}

// Tests LaunchTemplate as a side effect.
func TestAutoscalingGroupOSWithLaunchTemplate(t *testing.T) {
	stub := stubAWS(t)
	defer stub.Close()

	stubDescribeImages(stub, "ami-0227c65b90645ae0c", "RunInstances:0002")

	args := resources.AutoscalingGroup{
		LaunchTemplate: &resources.LaunchTemplate{AMI: "ami-0227c65b90645ae0c"},
	}
	resource := args.BuildResource()
	estimates := newEstimates(stub.ctx, t, resource)
	estimates.mustHave("operating_system", "windows")
}
