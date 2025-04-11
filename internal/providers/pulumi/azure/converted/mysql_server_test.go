package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestMySQLServer_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	opts := tftest.DefaultGoldenFileOptions()
	opts.CaptureLogs = true
	// ignore CLI - this has been removed from the latest provider
	opts.IgnoreCLI = true
	tftest.GoldenFileResourceTestsWithOpts(t, "mysql_server_test", opts)
}
