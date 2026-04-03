package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestSageMakerEndpointConfigurationGoldenFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// The second argument must match the folder name in testdata/
	tftest.GoldenFileResourceTestsWithOpts(t, "sagemaker_endpoint_configuration_test", &tftest.GoldenFileOptions{
		CaptureLogs: true,
	})
}