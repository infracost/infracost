package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/idem/idemtest"
)

func TestEC2TransitGatewayVpcAttachmentGoldenFile(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	opts := idemtest.DefaultGoldenFileOptions()
	opts.CaptureLogs = true
	idemtest.GoldenFileResourceTestsWithOpts(t, "ec2_transit_gateway_vpc_attachment_test", opts)
}
