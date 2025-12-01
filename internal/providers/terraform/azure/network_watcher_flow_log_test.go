package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestNetworkWatcherFlowLog(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tftest.GoldenFileResourceTestsWithOpts(t, "network_watcher_flow_log_test", &tftest.GoldenFileOptions{
		IgnoreCLI: true, // Azure deprecated NSG flow logs June 2025
	})
}
