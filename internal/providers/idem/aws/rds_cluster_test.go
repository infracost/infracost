package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/idem/idemtest"
)

func TestRDSClusterGoldenFile(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	opts := idemtest.DefaultGoldenFileOptions()
	opts.CaptureLogs = true
	idemtest.GoldenFileResourceTestsWithOpts(t, "rds_cluster_test", opts)
}
