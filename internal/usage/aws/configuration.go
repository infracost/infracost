//nolint:deadcode,unused
package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

type ctxKeyType struct{}

var ctxKey = &ctxKeyType{}

func getConfig(ctx context.Context, region string) (aws.Config, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
		// config.WithClientLogMode(aws.LogResponseWithBody),
	}

	if ctxOpts, ok := ctx.Value(ctxKey).([]func(*config.LoadOptions) error); ok {
		opts = append(opts, ctxOpts...)
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	return cfg, err
}
