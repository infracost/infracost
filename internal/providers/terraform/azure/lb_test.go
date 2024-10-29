package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAzureRMLoadBalancerGoldenFile(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	opts := tftest.DefaultGoldenFileOptions()

	// Skip CLI diff as it yields different results
	opts.IgnoreCLI = true

	tftest.GoldenFileResourceTestsWithOpts(t, "lb_test", opts)
}
