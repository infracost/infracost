//nolint:deadcode,unused
package aws

import (
	"context"
	"github.com/infracost/infracost/internal/usage"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

type ctxConfigOptsKeyType struct{}

var ctxConfigOptsKey = &ctxConfigOptsKeyType{}

var configMux sync.Mutex

func getConfig(ctx context.Context, region string) (aws.Config, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
		// config.WithClientLogMode(aws.LogRequestWithBody | aws.LogResponseWithBody),
	}

	if ctxOpts, ok := ctx.Value(ctxConfigOptsKey).([]func(*config.LoadOptions) error); ok {
		opts = append(opts, ctxOpts...)
	}

	// We want to set the OS env so that the AWS config loader picks up any AWS_*
	// env vars set in the Infracost config file. We use a mutex for this since
	// it's run in parallel and os.Setenv sets the global OS env for the process.
	configMux.Lock()
	defer configMux.Unlock()

	var oldEnv []string

	env, hasEnv := ctx.Value(usage.ContextEnv{}).(map[string]string)
	if hasEnv {
		oldEnv = os.Environ()
		for k, v := range env {
			os.Setenv(k, v)
		}
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	resetEnv(oldEnv)
	return cfg, err
}

func resetEnv(items []string) {
	os.Clearenv()
	for _, item := range items {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) == 2 {
			os.Setenv(parts[0], parts[1])
		}
	}
}
