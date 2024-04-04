package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/infracost/infracost/internal/logging"
)

// c.f. https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/billing-info-fields.html
var ec2UsageOperationMap = map[string]string{
	"RunInstances":      "linux", // Linux/UNIX
	"RunInstances:00g0": "rhel",  // Red Hat BYOL Linux
	"RunInstances:0010": "rhel",  // Red Hat Enterprise Linux
	"RunInstances:1010": "rhel",  // Red Hat Enterprise Linux with HA
	"RunInstances:1014": "rhel",  // Red Hat Enterprise Linux with SQL Server Standard and HA
	"RunInstances:1110": "rhel",  // Red Hat Enterprise Linux with SQL Server Enterprise and HA
	"RunInstances:0014": "rhel",  // Red Hat Enterprise Linux with SQL Server Standard
	"RunInstances:0210": "rhel",  // Red Hat Enterprise Linux with SQL Server Web
	"RunInstances:0110": "rhel",  // Red Hat Enterprise Linux with SQL Server Enterprise
	// "RunInstances:0100": "", // SQL Server Enterprise
	// "RunInstances:0004": "", // SQL Server Standard
	// "RunInstances:0200": "", // SQL Server Web
	"RunInstances:000g": "suse",    // SUSE Linux
	"RunInstances:0002": "windows", // Windows
	"RunInstances:0800": "windows", // Windows BYOL
}

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
	logging.Logger.Debug().Msgf("Querying AWS EC2 API: DescribeImages (region: %s, ImageIds: [%s])", region, ami)
	result, err := client.DescribeImages(ctx, &ec2.DescribeImagesInput{
		ImageIds: []string{ami},
	})
	if err != nil {
		return "", err
	} else if len(result.Images) == 0 {
		return "", nil
	}
	if result.Images[0].UsageOperation != nil {
		if uo, ok := ec2UsageOperationMap[*result.Images[0].UsageOperation]; ok {
			return uo, nil
		}
	}
	switch result.Images[0].Platform {
	case types.PlatformValuesWindows:
		return "windows", nil
	default:
		return "linux", nil
	}
}
