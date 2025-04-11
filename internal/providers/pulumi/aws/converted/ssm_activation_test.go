package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAwsSSMActivationfuncGoldenFile(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tftest.GoldenFileResourceTests(t, "ssm_activation_test")
}
