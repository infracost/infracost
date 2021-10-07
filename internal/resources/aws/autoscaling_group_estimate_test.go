package aws_test

import (
	"fmt"
	"testing"

	resources "github.com/infracost/infracost/internal/resources/aws"
)

func stubASGQuery(stub *stubbedAWS, name string, value float64) {
	var datapoints string

	if value > 0 {
		datapoints = fmt.Sprintf(`
			<member>
				<Average>%f</Average>
				<Unit>None</Unit>
				<Timestamp>1970-01-01T00:00:00Z</Timestamp>
			</member>
		`, value)
	}
	stub.WhenBody("Action=GetMetricStatistics&Dimensions.member.1.Name=AutoScalingGroupName&Dimensions.member.1.Value="+name).Then(200, `
		<GetMetricStatisticsResponse xmlns="http://monitoring.amazonaws.com/doc/2010-08-01/">
	  <GetMetricStatisticsResult>
	    <Datapoints>`+datapoints+`

	    </Datapoints>
	    <Label>GroupTotalInstances</Label>
	  </GetMetricStatisticsResult>
	</GetMetricStatisticsResponse>
	`)
}

func stubASGDescribe(stub *stubbedAWS, name string, count int64) {
	var instanceMembers string
	var groupMember string

	// shoddy stub: woefully incomplete compared to real response
	if count > 0 {
		for i := int64(0); i < count; i++ {
			instanceMembers += `
					<member></member>`
		}

		groupMember = `
			<member>
				<Instances>
				` + instanceMembers +
			`</Instances>
			</member>`
	}

	stub.WhenBody("Action=DescribeAutoScalingGroups&AutoScalingGroupNames.member.1="+name).Then(200, `
		<DescribeAutoScalingGroupsResponse xmlns="http://autoscaling.amazonaws.com/doc/2011-01-01/">
	  <DescribeAutoScalingGroupsResult>
	    <AutoScalingGroups>
		 	`+groupMember+`
	    </AutoScalingGroups>
	  </DescribeAutoScalingGroupsResult>
	  <ResponseMetadata>
	    <RequestId>8aea7709-e291-4e80-a89e-682e23109bb7</RequestId>
	  </ResponseMetadata>
	</DescribeAutoScalingGroupsResponse>
	`)
}

// Tests LaunchConfiguration as a side effect.
func TestAutoscalingGroupOSWithLaunchConfiguration(t *testing.T) {
	stub := stubAWS(t)
	defer stub.Close()

	stubDescribeImages(stub, "ami-0227c65b90645ae0c", "RunInstances:0002")
	stubASGQuery(stub, "deadbeef", 1) // don't care

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
	stubASGQuery(stub, "deadbeef", 1) // don't care

	args := resources.AutoscalingGroup{
		LaunchTemplate: &resources.LaunchTemplate{AMI: "ami-0227c65b90645ae0c"},
	}
	resource := args.BuildResource()
	estimates := newEstimates(stub.ctx, t, resource)
	estimates.mustHave("operating_system", "windows")
}

func TestAutoscalingGroupInstancesWithCloudWatch(t *testing.T) {
	stub := stubAWS(t)
	defer stub.Close()

	stubDescribeImages(stub, "ami-0227c65b90645ae0c", "RunInstances:0002") // don't care
	stubASGQuery(stub, "deadbeef", 3.14159)

	args := resources.AutoscalingGroup{
		Name: "deadbeef",
	}
	resource := args.BuildResource()
	estimates := newEstimates(stub.ctx, t, resource)
	estimates.mustHave("instances", int64(3))
}

func TestAutoscalingGroupInstancesWithoutCloudWatch(t *testing.T) {
	stub := stubAWS(t)
	defer stub.Close()

	stubDescribeImages(stub, "ami-0227c65b90645ae0c", "RunInstances:0002") // don't care
	stubASGQuery(stub, "deadbeef", 0)                                      // no results
	stubASGDescribe(stub, "deadbeef", 5)

	args := resources.AutoscalingGroup{
		Name: "deadbeef",
	}
	resource := args.BuildResource()
	estimates := newEstimates(stub.ctx, t, resource)
	estimates.mustHave("instances", int64(5))
}
