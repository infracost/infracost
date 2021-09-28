package aws_test

import (
	"testing"

	resources "github.com/infracost/infracost/internal/resources/aws"
)

func TestEKSNodeGroupOS(t *testing.T) {
	stub := stubAWS(t)
	defer stub.Close()

	args := resources.EKSNodeGroup{}
	resource := args.BuildResource()
	estimates := newEstimates(stub.ctx, t, resource)
	estimates.mustHave("operating_system", "linux")
}

// Tests LaunchTemplate as a side effect.
func TestEKSNodeGroupOSWithLaunchTemplate(t *testing.T) {
	stub := stubAWS(t)
	defer stub.Close()

	stubDescribeImages(stub, "ami-0227c65b90645ae0c", "RunInstances:0002")

	args := resources.EKSNodeGroup{
		LaunchTemplate: &resources.LaunchTemplate{AMI: "ami-0227c65b90645ae0c"},
	}
	resource := args.BuildResource()
	estimates := newEstimates(stub.ctx, t, resource)
	estimates.mustHave("operating_system", "windows")
}
