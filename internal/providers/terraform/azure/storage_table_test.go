package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAzureStorageTable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	opts := tftest.DefaultGoldenFileOptions()
	tftest.GoldenFileResourceTestsWithOpts(t, "storage_table_test", opts)
}
