package aws

import (
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

func lambdaGetInvocations(region string, fn string) float64 {
	namespace := "AWS/Lambda"
	metric := "Invocations"
	stats, err := sdkGetMonthlyStats(sdkStatsRequest{
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
		sdkWarn(namespace, metric, fn, err)
	} else if len(stats.Datapoints) > 0 {
		return *stats.Datapoints[0].Sum
	}
	return 0
}

func lambdaGetDuration(region string, fn string) float64 {
	namespace := "AWS/Lambda"
	metric := "Duration"
	stats, err := sdkGetMonthlyStats(sdkStatsRequest{
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
		sdkWarn(namespace, metric, fn, err)
		return 0
	} else if len(stats.Datapoints) == 0 {
		return 0
	}
	return *stats.Datapoints[0].Average
}
