package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func ec2NewClient(ctx context.Context, region string) (*ec2.Client, error) {
	cfg, err := getConfig(ctx, region)
	if err != nil {
		return nil, err
	}
	return ec2.NewFromConfig(cfg), nil
}

func EC2DescribeOS(ctx context.Context, region string, ami string) (string, error) {
	client, err := ec2NewClient(ctx, region)
	if err != nil {
		return "", err
	}
	result, err := client.DescribeImages(ctx, &ec2.DescribeImagesInput{
		ImageIds: []string{ami},
	})
	if err != nil {
		return "", err
	} else if len(result.Images) == 0 {
		return "", nil
	}
	switch result.Images[0].Platform {
	case types.PlatformValuesWindows:
		return "windows", nil
	default:
		return "", nil
	}
}
