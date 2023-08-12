package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAzureRMLinuxAppFunctionGoldenFile(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	opts := tftest.DefaultGoldenFileOptions()
	// ignore the Terraform Plan JSON as the Terraform cannot traverse the each.value references correctly
	// meaning that the HCL provider is more accurate here.
	opts.IgnorePlanJSON = true
	tftest.GoldenFileResourceTestsWithOpts(t, "function_linux_app_test", opts)
}
