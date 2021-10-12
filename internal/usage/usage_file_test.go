package usage_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestUsageFile(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tftest.EnsurePluginsInstalled()
	tftest.GoldenFileUsageSyncTest(t, "usage_file")
}

func TestUsageFileNoUsage(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tftest.EnsurePluginsInstalled()
	tftest.GoldenFileUsageSyncTest(t, "usage_file_no_usage")
}

func TestUsageFileEmpty(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tftest.EnsurePluginsInstalled()
	tftest.GoldenFileUsageSyncTest(t, "usage_file_empty")
}
