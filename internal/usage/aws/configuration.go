//nolint:deadcode,unused
package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

type ctxConfigOptsKeyType struct{}

var ctxConfigOptsKey = &ctxConfigOptsKeyType{}

func getConfig(ctx context.Context, region string) (aws.Config, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
		// config.WithClientLogMode(aws.LogRequestWithBody | aws.LogResponseWithBody),
	}

	if ctxOpts, ok := ctx.Value(ctxConfigOptsKey).([]func(*config.LoadOptions) error); ok {
		opts = append(opts, ctxOpts...)
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	return cfg, err
}
