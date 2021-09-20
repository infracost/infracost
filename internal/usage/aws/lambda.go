//nolint:deadcode,unused
package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

func LambdaGetInvocations(ctx context.Context, region string, fn string) (float64, error) {
	namespace := "AWS/Lambda"
	metric := "Invocations"
	stats, err := cloudwatchGetMonthlyStats(ctx, statsRequest{
		region:    region,
		namespace: namespace,
		metric:    metric,
		statistic: types.StatisticSum,
		unit:      types.StandardUnitCount,
		dimensions: map[string]string{
			"FunctionName": fn,
		},
	})
	if err != nil {
		return 0, err
	} else if len(stats.Datapoints) > 0 {
		return *stats.Datapoints[0].Sum, nil
	}
	return 0, nil
}

func LambdaGetDurationAvg(ctx context.Context, region string, fn string) (float64, error) {
	namespace := "AWS/Lambda"
	metric := "Duration"
	stats, err := cloudwatchGetMonthlyStats(ctx, statsRequest{
		region:    region,
		namespace: namespace,
		metric:    metric,
		statistic: types.StatisticAverage,
		unit:      types.StandardUnitMilliseconds,
		dimensions: map[string]string{
			"FunctionName": fn,
		},
	})
	if err != nil {
		return 0, err
	} else if len(stats.Datapoints) == 0 {
		return 0, nil
	}
	return *stats.Datapoints[0].Average, nil
}
