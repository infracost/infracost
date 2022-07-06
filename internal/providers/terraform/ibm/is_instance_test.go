package ibm_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestIsInstance(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	//opts := tftest.DefaultGoldenFileOptions()
	//opts.CaptureLogs = true
	//tftest.GoldenFileResourceTestsWithOpts(t, "is_instance_test", opts)

	tftest.GoldenFileResourceTests(t, "is_instance_test")
}
