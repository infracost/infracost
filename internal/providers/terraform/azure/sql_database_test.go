package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestSQLDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	opts := tftest.DefaultGoldenFileOptions()
	opts.CaptureLogs = true
	// ignore CLI - this has been removed from the latest provider
	opts.IgnoreCLI = true
	tftest.GoldenFileResourceTestsWithOpts(t, "sql_database_test", opts)
}
