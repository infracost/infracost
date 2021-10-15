//nolint:deadcode,unused,varcheck
package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

const statAvg = types.StatisticAverage
const statSum = types.StatisticSum

const unitCount = types.StandardUnitCount

func cloudwatchNewClient(ctx context.Context, region string) (*cloudwatch.Client, error) {
	cfg, err := getConfig(ctx, region)
	if err != nil {
		return nil, err
	}

	return cloudwatch.NewFromConfig(cfg), nil
}

type statsRequest struct {
	region     string
	namespace  string
	metric     string
	dimensions map[string]string
	statistic  types.Statistic
	unit       types.StandardUnit
}

func cloudwatchGetMonthlyStats(ctx context.Context, req statsRequest) (*cloudwatch.GetMetricStatisticsOutput, error) {
	client, err := cloudwatchNewClient(ctx, req.region)
	if err != nil {
		return nil, err
	}
	dim := make([]types.Dimension, 0, len(req.dimensions))
	for k, v := range req.dimensions {
		dim = append(dim, types.Dimension{
			Name:  strPtr(k),
			Value: strPtr(v),
		})
	}

	return client.GetMetricStatistics(ctx, &cloudwatch.GetMetricStatisticsInput{
		Namespace:  strPtr(req.namespace),
		MetricName: strPtr(req.metric),
		StartTime:  aws.Time(time.Now().Add(-timeMonth)),
		EndTime:    aws.Time(time.Now()),
		Period:     int32Ptr(60 * 60 * 24 * 30),
		Statistics: []types.Statistic{req.statistic},
		Unit:       req.unit,
		Dimensions: dim,
	})
}
