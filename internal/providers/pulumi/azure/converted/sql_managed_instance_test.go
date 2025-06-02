package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestSQLManagedInstance(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	opts := tftest.DefaultGoldenFileOptions()
	// This resource is removed from the latest provider
	opts.IgnoreCLI = true

	tftest.GoldenFileResourceTestsWithOpts(t, "sql_managed_instance_test", opts)
}
