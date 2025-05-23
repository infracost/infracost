package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAzureRMLoadBalancerRuleGoldenFile(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	opts := tftest.DefaultGoldenFileOptions()
	// Skip CLI diff as it yields different results
	opts.IgnoreCLI = true

	tftest.GoldenFileResourceTestsWithOpts(t, "lb_rule_test", opts)
}

func TestAzureRMLoadBalancerRuleV2GoldenFile(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	opts := tftest.DefaultGoldenFileOptions()
	// Skip CLI diff as it yields different results
	opts.IgnoreCLI = true

	tftest.GoldenFileResourceTestsWithOpts(t, "lb_rule_v2_test", opts)
}
