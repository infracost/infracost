package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestComputeInstanceGoldenFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	opts := tftest.DefaultGoldenFileOptions()
	opts.CaptureLogs = true
	tftest.GoldenFileResourceTestsWithOpts(t, "compute_instance_test", opts)
}
