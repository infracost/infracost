package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAppConfiguration(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tftest.GoldenFileResourceTestsWithOpts(t, "app_configuration_test", &tftest.GoldenFileOptions{
		Currency:    "USD",
		CaptureLogs: false,
		// ignore the CLI as this throws errors for the test case with an empty sku we
		// want to test for this case as well as this is a valid case for vscode cli
		// users who are yet to run the terraform plan/apply step.
		IgnoreCLI: true,
	})
}
