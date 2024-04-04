package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"

	"github.com/infracost/infracost/internal/logging"
)

func autoscalingNewClient(ctx context.Context, region string) (*autoscaling.Client, error) {
	cfg, err := getConfig(ctx, region)
	if err != nil {
		return nil, err
	}
	return autoscaling.NewFromConfig(cfg), nil
}

// AutoscalingGetInstanceCount uses various techniques to estimate number of instances in an AutoScaling group.
// In order of preference:
//  1. CloudWatch monthly average
//  2. Instantaneous count right now
//  3. Mean of min-size and max-size
func AutoscalingGetInstanceCount(ctx context.Context, region string, name string) (float64, error) {
	logging.Logger.Debug().Msgf("Querying AWS CloudWatch: AWS/AutoScaling GroupTotalInstances (region: %s, AutoScalingGroupName: %s)", region, name)
	stats, err := cloudwatchGetMonthlyStats(ctx, statsRequest{
		region:     region,
		namespace:  "AWS/AutoScaling",
		metric:     "GroupTotalInstances",
		dimensions: map[string]string{"AutoScalingGroupName": name},
		statistic:  statAvg,
	})
	if err != nil {
		return 0, err
	}
	if len(stats.Datapoints) > 0 {
		return *stats.Datapoints[0].Average, nil
	}

	client, err := autoscalingNewClient(ctx, region)
	if err != nil {
		return 0, err
	}
	logging.Logger.Debug().Msgf("Querying AWS EC2 API: DescribeAutoScalingGroups(region: %s, AutoScalingGroupNames: [%s])", region, name)
	resp, err := client.DescribeAutoScalingGroups(ctx, &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{name},
	})
	if err != nil {
		return 0, err
	}
	if len(resp.AutoScalingGroups) == 0 {
		return 0, nil
	}
	asg := resp.AutoScalingGroups[0]
	now := len(asg.Instances)
	if now > 0 {
		return float64(now), nil
	}
	return 0, nil
}
