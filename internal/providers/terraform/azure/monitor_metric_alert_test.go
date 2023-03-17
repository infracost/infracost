package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestMonitorMetricAlert(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	opts := tftest.DefaultGoldenFileOptions()
	opts.CaptureLogs = true
	tftest.GoldenFileResourceTestsWithOpts(t, "monitor_metric_alert_test", opts)
}
