//nolint:deadcode,unused
package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

type testResolver struct {
	URL string
}

func (tr *testResolver) ResolveEndpoint(service, region string) (aws.Endpoint, error) {
	return aws.Endpoint{
		URL: tr.URL,
	}, nil
}

type testCredentials struct {
}

func (tc *testCredentials) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return aws.Credentials{AccessKeyID: "AKIA0123456789", SecretAccessKey: "opensesame"}, nil
}

func WithTestEndpoint(ctx context.Context, url string) context.Context {
	resolver := &testResolver{URL: url}
	opts := []func(*config.LoadOptions) error{
		config.WithEndpointResolver(resolver),
		config.WithCredentialsProvider(&testCredentials{}),
		// config.WithClientLogMode(aws.LogResponseWithBody),
	}
	return context.WithValue(ctx, ctxKey, opts)
}
