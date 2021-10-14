package aws_test

import (
	"fmt"
	"strings"
	"testing"

	resources "github.com/infracost/infracost/internal/resources/aws"
	"github.com/stretchr/testify/assert"
)

func stubEKSDescribeNodegroup(stub *stubbedAWS, clusterName, nodeGroupName string, asgNames []string) {
	asgMembers := []string{}
	for _, asgName := range asgNames {
		asgMembers = append(asgMembers, fmt.Sprintf(`{"name": "%s"}`, asgName))
	}

	stub.WhenFullPath(fmt.Sprintf("/clusters/%s/node-groups/%s", clusterName, nodeGroupName)).Then(200, fmt.Sprintf(`
	{
		"nodegroup": {
			"nodegroupName": "%s",
			"clusterName": "%s",
			"resources": {
				"autoScalingGroups": [%s]
			}
		}
	}
	`, nodeGroupName, clusterName, strings.Join(asgMembers, ",")))
}

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

	stubEC2DescribeImages(stub, "ami-0227c65b90645ae0c", "RunInstances:0002")

	args := resources.EKSNodeGroup{
		LaunchTemplate: &resources.LaunchTemplate{AMI: "ami-0227c65b90645ae0c"},
	}
	resource := args.BuildResource()
	estimates := newEstimates(stub.ctx, t, resource)
	assert.Equal(t, "windows", estimates.usage["operating_system"])
}

func TestEKSNodeGroupInstancesWithCloudWatch(t *testing.T) {
	stub := stubAWS(t)
	defer stub.Close()

	stubEC2DescribeImages(stub, "ami-0227c65b90645ae0c", "RunInstances:0002") // don't care
	stubEKSDescribeNodegroup(stub, "eks-cluster-name", "eks-node-group-name", []string{"asg-1", "asg-2"})
	stubCloudWatchASGQuery(stub, "asg-1", 3.14159)
	stubCloudWatchASGQuery(stub, "asg-2", 2.71828)

	args := resources.EKSNodeGroup{
		Name:        "eks-node-group-name",
		ClusterName: "eks-cluster-name",
	}
	resource := args.BuildResource()
	estimates := newEstimates(stub.ctx, t, resource)
	assert.Equal(t, int64(6), estimates.usage["instances"])
}

func TestEKSNodeGroupInstancesWithoutCloudWatch(t *testing.T) {
	stub := stubAWS(t)
	defer stub.Close()

	stubEC2DescribeImages(stub, "ami-0227c65b90645ae0c", "RunInstances:0002") // don't care
	stubEKSDescribeNodegroup(stub, "eks-cluster-name", "eks-node-group-name", []string{"asg-1", "asg-2"})
	stubCloudWatchASGQuery(stub, "asg-1", 0) // no results
	stubCloudWatchASGQuery(stub, "asg-2", 0) // no results
	stubEC2DescribeAutoscalingGroups(stub, "asg-1", 2)
	stubEC2DescribeAutoscalingGroups(stub, "asg-2", 3)

	args := resources.EKSNodeGroup{
		Name:        "eks-node-group-name",
		ClusterName: "eks-cluster-name",
	}
	resource := args.BuildResource()
	estimates := newEstimates(stub.ctx, t, resource)
	assert.Equal(t, int64(5), estimates.usage["instances"])
}
