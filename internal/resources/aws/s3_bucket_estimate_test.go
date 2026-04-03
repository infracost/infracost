package aws_test

import (
	"fmt"
	"testing"

	resources "github.com/infracost/infracost/internal/resources/aws"
	"github.com/stretchr/testify/assert"
)

func stubListBucketMetricsConfigurations(stub *stubbedAWS) {
	stub.WhenFullPath("/test-bucket?metrics=&x-id=ListBucketMetricsConfigurations").Then(200, `
		<ListMetricsConfigurationsResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
			<MetricsConfiguration>
				<Filter>
					<Prefix>test-prefix</Prefix>
					<Id>withPrefix</Id>
				</Filter>
			</MetricsConfiguration>
			<MetricsConfiguration>
				<Id>infracost</Id>
			</MetricsConfiguration>
			<IsTruncated>false</IsTruncated>
		</ListMetricsConfigurationsResult>`)
}

func stubListBucketMetricsConfigurationsNoMatching(stub *stubbedAWS) {
	stub.WhenFullPath("/test-bucket?metrics=&x-id=ListBucketMetricsConfigurations").Then(200, `
		<ListMetricsConfigurationsResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
			<MetricsConfiguration>
				<Filter>
					<Prefix>test-prefix</Prefix>
					<Id>withPrefix</Id>
				</Filter>
			</MetricsConfiguration>
			<IsTruncated>false</IsTruncated>
		</ListMetricsConfigurationsResult>`)
}

func stubStorageClassBytes(stub *stubbedAWS, storageClass string, bytes int) {
	stub.WhenBody(fmt.Sprintf("Value=%s", storageClass), "MetricName=BucketSizeBytes", "Statistics.member.1=Average", "Unit=Bytes").Then(200, fmt.Sprintf(`
	<GetMetricStatisticsResponse xmlns="http://monitoring.amazonaws.com/doc/2010-08-01/">
		<GetMetricStatisticsResult>
			<Label>BucketSizeBytes</Label>
			<Datapoints>
				<member>
					<Unit>Bytes</Unit>
					<Average>%d</Average>
					<Timestamp>1970-01-01T00:00:00Z</Timestamp>
				</member>
			</Datapoints>
		</GetMetricStatisticsResult>
	</GetMetricStatisticsResponse>`, bytes))
}

func stubRequestCounts(stub *stubbedAWS, metric string, count int) {
	stub.WhenBody("Value=infracost", fmt.Sprintf("MetricName=%s", metric), "Statistics.member.1=Sum", "Unit=Count").Then(200, fmt.Sprintf(`
	<GetMetricStatisticsResponse xmlns="http://monitoring.amazonaws.com/doc/2010-08-01/">
		<GetMetricStatisticsResult>
			<Label>%s</Label>
			<Datapoints>
				<member>
					<Unit>Count</Unit>
					<Sum>%d</Sum>
					<Timestamp>1970-01-01T00:00:00Z</Timestamp>
				</member>
			</Datapoints>
		</GetMetricStatisticsResult>
	</GetMetricStatisticsResponse>`, metric, count))
}

func stubDataBytes(stub *stubbedAWS, metric string, bytes int) {
	stub.WhenBody("Value=infracost", fmt.Sprintf("MetricName=%s", metric), "Statistics.member.1=Sum", "Unit=Bytes").Then(200, fmt.Sprintf(`
	<GetMetricStatisticsResponse xmlns="http://monitoring.amazonaws.com/doc/2010-08-01/">
		<GetMetricStatisticsResult>
			<Label>%s</Label>
			<Datapoints>
				<member>
					<Unit>Bytes</Unit>
					<Sum>%d</Sum>
					<Timestamp>1970-01-01T00:00:00Z</Timestamp>
				</member>
			</Datapoints>
		</GetMetricStatisticsResult>
	</GetMetricStatisticsResponse>`, metric, bytes))
}

func TestS3Bucket(t *testing.T) {
	stub := stubAWS(t)
	defer stub.Close()

	stubListBucketMetricsConfigurations(stub)

	storageClassBytes := map[string]int{
		"StandardStorage":              2100000000,
		"IntelligentTieringFAStorage":  2200000000,
		"IntelligentTieringIAStorage":  2300000000,
		"IntelligentTieringAAStorage":  2400000000,
		"IntelligentTieringDAAStorage": 2500000000,
		"StandardIAStorage":            2600000000,
		"OneZoneIAStorage":             2700000000,
		"GlacierStorage":               2800000000,
		"DeepArchiveStorage":           0, // This should not appear in estimates.usages
	}

	for storageClass, bytes := range storageClassBytes {
		stubStorageClassBytes(stub, storageClass, bytes)
	}

	requestCounts := map[string]int{
		"PutRequests":    100,
		"PostRequests":   200,
		"ListRequests":   300,
		"GetRequests":    400,
		"HeadRequests":   500,
		"SelectRequests": 600,
	}

	for metric, count := range requestCounts {
		stubRequestCounts(stub, metric, count)
	}

	dataBytes := map[string]int{
		"SelectBytesScanned":  1100000000,
		"SelectBytesReturned": 1200000000,
	}

	for metric, bytes := range dataBytes {
		stubDataBytes(stub, metric, bytes)
	}

	args := resources.S3Bucket{
		Name:   "test-bucket",
		Region: "us-east-1",
	}
	resource := args.BuildResource()
	estimates := newEstimates(stub.ctx, t, resource)

	assert.Equal(t, map[string]any{
		"standard": map[string]any{
			"storage_gb":                      2.1,
			"monthly_tier_1_requests":         int64(600),
			"monthly_tier_2_requests":         int64(1500),
			"monthly_select_data_scanned_gb":  1.1,
			"monthly_select_data_returned_gb": 1.2,
		},
		"intelligent_tiering": map[string]any{
			"frequent_access_storage_gb":     2.2,
			"infrequent_access_storage_gb":   2.3,
			"archive_access_storage_gb":      2.4,
			"deep_archive_access_storage_gb": 2.5,
		},
		"standard_infrequent_access": map[string]any{
			"storage_gb": 2.6,
		},
		"one_zone_infrequent_access": map[string]any{
			"storage_gb": 2.7,
		},
		"glacier_flexible_retrieval": map[string]any{
			"storage_gb": 2.8,
		},
	}, estimates.usage)
}

func TestS3BucketNoFilter(t *testing.T) {
	stub := stubAWS(t)
	defer stub.Close()

	stubListBucketMetricsConfigurationsNoMatching(stub)

	storageClassBytes := map[string]int{
		"StandardStorage":              2100000000,
		"IntelligentTieringFAStorage":  2200000000,
		"IntelligentTieringIAStorage":  2300000000,
		"IntelligentTieringAAStorage":  2400000000,
		"IntelligentTieringDAAStorage": 2500000000,
		"StandardIAStorage":            2600000000,
		"OneZoneIAStorage":             2700000000,
		"GlacierStorage":               2800000000,
		"DeepArchiveStorage":           2900000000,
	}

	for storageClass, bytes := range storageClassBytes {
		stubStorageClassBytes(stub, storageClass, bytes)
	}

	args := resources.S3Bucket{
		Name:   "test-bucket",
		Region: "us-east-1",
	}
	resource := args.BuildResource()
	estimates := newEstimates(stub.ctx, t, resource)

	assert.Equal(t, map[string]any{
		"storage_gb": 2.1,
	}, estimates.usage["standard"])
}

func TestS3BucketNoStandard(t *testing.T) {
	stub := stubAWS(t)
	defer stub.Close()

	stubListBucketMetricsConfigurations(stub)

	storageClassBytes := map[string]int{
		"StandardStorage":              0,
		"IntelligentTieringFAStorage":  2200000000,
		"IntelligentTieringIAStorage":  0,
		"IntelligentTieringAAStorage":  0,
		"IntelligentTieringDAAStorage": 0,
		"StandardIAStorage":            0,
		"OneZoneIAStorage":             0,
		"GlacierStorage":               0,
		"DeepArchiveStorage":           0,
	}

	for storageClass, bytes := range storageClassBytes {
		stubStorageClassBytes(stub, storageClass, bytes)
	}

	requestCounts := map[string]int{
		"PutRequests":    100,
		"PostRequests":   200,
		"ListRequests":   300,
		"GetRequests":    400,
		"HeadRequests":   500,
		"SelectRequests": 600,
	}

	for metric, count := range requestCounts {
		stubRequestCounts(stub, metric, count)
	}

	dataBytes := map[string]int{
		"SelectBytesScanned":  1100000000,
		"SelectBytesReturned": 1200000000,
	}

	for metric, bytes := range dataBytes {
		stubDataBytes(stub, metric, bytes)
	}

	args := resources.S3Bucket{
		Name:   "test-bucket",
		Region: "us-east-1",
	}
	resource := args.BuildResource()
	estimates := newEstimates(stub.ctx, t, resource)

	assert.Equal(t, map[string]any{
		"standard": map[string]any{
			"storage_gb":                      float64(0),
			"monthly_tier_1_requests":         int64(600),
			"monthly_tier_2_requests":         int64(1500),
			"monthly_select_data_scanned_gb":  1.1,
			"monthly_select_data_returned_gb": 1.2,
		},
		"intelligent_tiering": map[string]any{
			"frequent_access_storage_gb": 2.2,
		},
	}, estimates.usage)
}
