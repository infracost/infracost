package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/arm/armtest"
)

func TestAzureRMManagedDiskGoldenFile(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	armtest.GoldenFileResourceTests(t, "managed_disk_test")
}
