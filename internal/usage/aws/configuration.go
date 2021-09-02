package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/infracost/infracost/internal/config"
)

// ConfigureEstimation caches AWS SDK configuration in the project context.
func ConfigureEstimation(ctx *config.ProjectContext) error {
	return nil
}

func getConfig(ctx *config.ProjectContext, region string) (aws.Config, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion(region))
	return cfg, err
}
