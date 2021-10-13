package aws_test

import (
	"testing"

	resources "github.com/infracost/infracost/internal/resources/aws"
	"github.com/stretchr/testify/assert"
)

func TestEKSNodeGroupOS(t *testing.T) {
	stub := stubAWS(t)
	defer stub.Close()

	args := resources.EKSNodeGroup{}
	resource := args.BuildResource()
	estimates := newEstimates(stub.ctx, t, resource)
	assert.Equal(t, "linux", estimates.usage["operating_system"])
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
	assert.Equal(t, "windows", estimates.usage["operating_system"])
}
