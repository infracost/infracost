//nolint:deadcode,unused
package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/infracost/infracost/internal/logging"
)

type ctxS3ConfigOptsKeyType struct{}

var ctxS3ConfigOptsKey = &ctxS3ConfigOptsKeyType{}

func s3NewClient(ctx context.Context, region string) (*s3.Client, error) {
	cfg, err := getConfig(ctx, region)
	if err != nil {
		return nil, err
	}

	opts := func(o *s3.Options) {}
	if ctxS3ConfigOpts, ok := ctx.Value(ctxS3ConfigOptsKey).(func(o *s3.Options)); ok {
		opts = ctxS3ConfigOpts
	}

	return s3.NewFromConfig(cfg, opts), nil
}

func S3FindMetricsFilter(ctx context.Context, region string, bucket string) (string, error) {
	client, err := s3NewClient(ctx, region)
	if err != nil {
		return "", err
	}
	logging.Logger.Debug().Msgf("Querying AWS S3 API: ListBucketMetricsConfigurations(region: %s, Bucket: %s)", region, bucket)
	result, err := client.ListBucketMetricsConfigurations(ctx, &s3.ListBucketMetricsConfigurationsInput{
		Bucket: strPtr(bucket),
	})

	if err != nil {
		return "", err
	}
	for _, config := range result.MetricsConfigurationList {
		if config.Filter == nil {
			return *config.Id, nil
		}
	}
	return "", nil
}

func S3GetBucketSizeBytes(ctx context.Context, region string, bucket string, storageType string) (float64, error) {
	logging.Logger.Debug().Msgf("Querying AWS CloudWatch: AWS/S3 BucketSizeBytes (region: %s, BucketName: %s, StorageType: %s)", region, bucket, storageType)
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
		return 0, err
	} else if len(stats.Datapoints) == 0 {
		return 0, nil
	}
	return *stats.Datapoints[0].Average, nil
}

func S3GetBucketRequests(ctx context.Context, region string, bucket string, filterName string, metrics []string) (int64, error) {
	count := int64(0)
	for _, metric := range metrics {
		logging.Logger.Debug().Msgf("Querying AWS CloudWatch: AWS/S3 %s (region: %s, BucketName: %s, FilterId: %s)", metric, region, bucket, filterName)
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
			return 0, err
		} else if len(stats.Datapoints) > 0 {
			count += int64(*stats.Datapoints[0].Sum)
		}
	}
	return count, nil
}

func S3GetBucketDataBytes(ctx context.Context, region string, bucket string, filterName string, metric string) (float64, error) {
	logging.Logger.Debug().Msgf("Querying AWS CloudWatch: AWS/S3 %s (region: %s, BucketName: %s, FilterId: %s)", metric, region, bucket, filterName)
	stats, err := cloudwatchGetMonthlyStats(ctx, statsRequest{
		region:    region,
		namespace: "AWS/S3",
		metric:    metric,
		statistic: types.StatisticSum,
		unit:      types.StandardUnitBytes,
		dimensions: map[string]string{
			"BucketName": bucket,
			"FilterId":   filterName,
		},
	})
	if err != nil {
		return 0, err
	} else if len(stats.Datapoints) == 0 {
		return 0, nil
	}
	return *stats.Datapoints[0].Sum, nil
}
