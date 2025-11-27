package azure_test

import (
	"testing"
	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAADB2CDirectoryGoldenFile(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	opts := tftest.DefaultGoldenFileOptions()
	tftest.GoldenFileResourceTestsWithOpts(t, "aadb2c_directory_test", opts)
} 