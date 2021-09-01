package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

func cloudwatchNewClient(region string) (*cloudwatch.Client, error) {
	config, err := sdkNewConfig(region)
	if err != nil {
		return nil, err
	}
	return cloudwatch.NewFromConfig(config), nil
}

type statsRequest struct {
	region     string
	namespace  string
	metric     string
	dimensions map[string]string
	statistic  types.Statistic
	unit       types.StandardUnit
}

func cloudwatchGetMonthlyStats(req statsRequest) (*cloudwatch.GetMetricStatisticsOutput, error) {
	client, err := sdkNewCloudWatchClient(req.region)
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
	return client.GetMetricStatistics(context.TODO(), &cloudwatch.GetMetricStatisticsInput{
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
