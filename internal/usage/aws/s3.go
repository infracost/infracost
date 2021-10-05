//nolint:deadcode,unused
package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func s3NewClient(ctx context.Context, region string) (*s3.Client, error) {
	cfg, err := getConfig(ctx, region)
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(cfg), nil
}

func s3FindMetricsFilter(ctx context.Context, region string, bucket string) string {
	client, err := s3NewClient(ctx, region)
	if err != nil {
		sdkWarn("S3", "requests", bucket, err)
		return ""
	}
	result, err := client.ListBucketMetricsConfigurations(ctx, &s3.ListBucketMetricsConfigurationsInput{
		Bucket: strPtr(bucket),
	})
	if err != nil {
		sdkWarn("S3", "requests", bucket, err)
		return ""
	}
	for _, config := range result.MetricsConfigurationList {
		if config.Filter == nil {
			return *config.Id
		}
	}
	return ""
}

func s3GetBucketSizeBytes(ctx context.Context, region string, bucket string, storageType string) float64 {
	stats, err := cloudwatchGetMonthlyStats(ctx, statsRequest{
		region:    region,
		namespace: "AWS/S3",
		metric:    "BucketSizeBytes",
		statistic: types.StatisticAverage,
		unit:      types.StandardUnitBytes,
		dimensions: map[string]string{
			"BucketName":  bucket,
			"StorageType": storageType,
		},
	})
	if err != nil {
		sdkWarn("S3", storageType, bucket, err)
		return 0
	} else if len(stats.Datapoints) == 0 {
		return 0
	}
	return *stats.Datapoints[0].Average
}

func s3GetBucketRequests(ctx context.Context, region string, bucket string, filterName string, metrics []string) float64 {
	count := float64(0)
	for _, metric := range metrics {
		stats, err := cloudwatchGetMonthlyStats(ctx, statsRequest{
			region:    region,
			namespace: "AWS/S3",
			metric:    metric,
			statistic: types.StatisticSum,
			unit:      types.StandardUnitCount,
			dimensions: map[string]string{
				"BucketName": bucket,
				"FilterId":   filterName,
			},
		})
		if err != nil {
			desc := fmt.Sprintf("%s per filter %s", metric, filterName)
			sdkWarn("S3", desc, bucket, err)
		} else if len(stats.Datapoints) > 0 {
			count += *stats.Datapoints[0].Sum
		}
	}
	return count
}
