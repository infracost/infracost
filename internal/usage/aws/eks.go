//nolint:deadcode,unused
package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	log "github.com/sirupsen/logrus"
)

func eksNewClient(ctx context.Context, region string) (*eks.Client, error) {
	cfg, err := getConfig(ctx, region)
	if err != nil {
		return nil, err
	}
	return eks.NewFromConfig(cfg), nil
}

func EKSGetNodeGroupAutoscalingGroups(ctx context.Context, region string, clusterName string, nodeGroupName string) ([]string, error) {
	client, err := eksNewClient(ctx, region)
	if err != nil {
		return []string{}, err
	}

	log.Debugf("Querying AWS EKS API: DescribeNodegroup(region: %s, ClusterName: %s, NodegroupName: %s)", region, clusterName, nodeGroupName)
	result, err := client.DescribeNodegroup(ctx, &eks.DescribeNodegroupInput{
		ClusterName:   strPtr(clusterName),
		NodegroupName: strPtr(nodeGroupName),
	})
	if err != nil {
		return []string{}, err
	}

	asgNames := make([]string, 0, len(result.Nodegroup.Resources.AutoScalingGroups))
	for _, asg := range result.Nodegroup.Resources.AutoScalingGroups {
		asgNames = append(asgNames, aws.ToString(asg.Name))
	}

	return asgNames, nil
}
