//nolint:deadcode,unused
package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type testCredentials struct {
}

func (tc *testCredentials) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return aws.Credentials{AccessKeyID: "AKIA0123456789", SecretAccessKey: "opensesame"}, nil
}

func WithTestEndpoint(ctx context.Context, url string) context.Context {
	opts := []func(*config.LoadOptions) error{
		config.WithBaseEndpoint(url),
		config.WithCredentialsProvider(&testCredentials{}),
		// config.WithClientLogMode(aws.LogRequestWithBody | aws.LogResponseWithBody),
	}
	ctx = context.WithValue(ctx, ctxConfigOptsKey, opts)

	s3Opts := func(o *s3.Options) {
		// We need this so the SDK doesn't use a subdomain for its requests
		o.UsePathStyle = true
	}

	ctx = context.WithValue(ctx, ctxS3ConfigOptsKey, s3Opts)
	return ctx
}
