package aws_test

import (
	"testing"

	resources "github.com/infracost/infracost/internal/resources/aws"
)

func TestLambda(t *testing.T) {
	stub := stubAWS(t)
	defer stub.Close()

	stub.WhenBody("GetMetricStatistics", "MetricName=Invocations", "Statistics.member.1=Sum").Then(200, `
		<GetMetricStatisticsResponse xmlns="http://monitoring.amazonaws.com/doc/2010-08-01/">
		  <GetMetricStatisticsResult>
		    <Datapoints>
				<member>
		        <Unit>Count</Unit>
		        <Sum>1234.0</Sum>
		        <Timestamp>1970-01-01T00:00:00Z</Timestamp>
		      </member>
		    </Datapoints>
		    <Label>Invocations</Label>
		  </GetMetricStatisticsResult>
		  <ResponseMetadata>
		    <RequestId>00000000-0000-0000-0000-000000000000</RequestId>
		  </ResponseMetadata>
		</GetMetricStatisticsResponse>
	`)
	stub.WhenBody("GetMetricStatistics", "MetricName=Duration", "Statistics.member.1=Average").Then(200, `
		<GetMetricStatisticsResponse xmlns="http://monitoring.amazonaws.com/doc/2010-08-01/">
		  <GetMetricStatisticsResult>
		    <Datapoints>
		      <member>
		        <Average>5678.9</Average>
		        <Unit>Milliseconds</Unit>
		        <Timestamp>1970-01-01T00:00:00Z</Timestamp>
		      </member>
		    </Datapoints>
		    <Label>Duration</Label>
		  </GetMetricStatisticsResult>
		  <ResponseMetadata>
		    <RequestId>00000000-0000-0000-0000-000000000000</RequestId>
		  </ResponseMetadata>
		</GetMetricStatisticsResponse>
	`)

	args := &resources.LambdaFunction{}
	resource := args.BuildResource()
	estimates := newEstimates(stub.ctx, t, resource)
	estimates.mustHave("monthly_requests", int64(1234))
	estimates.mustHave("request_duration_ms", int64(5679))
}
