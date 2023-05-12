package aws_test

import (
	"github.com/infracost/infracost/internal/providers/idem/idemtest"
	"testing"
)

func TestACMCertificateGoldenFile(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	opts := idemtest.DefaultGoldenFileOptions()
	opts.CaptureLogs = true
	idemtest.GoldenFileResourceTestsWithOpts(t, "acm_certificate_test", opts)
}
