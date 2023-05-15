package aws_test

import (
	"github.com/infracost/infracost/internal/providers/idem/idemtest"
	"testing"
)

func TestEBSSnapshotGoldenFile(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	opts := idemtest.DefaultGoldenFileOptions()
	opts.CaptureLogs = true
	idemtest.GoldenFileResourceTestsWithOpts(t, "ebs_snapshot_test", opts)
}
