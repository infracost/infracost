package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestDataFactoryIntegrationRuntimeManaged(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	opts := tftest.DefaultGoldenFileOptions()
	// Ignore the CLI because the resource has been removed from the provider
	opts.IgnoreCLI = true

	tftest.GoldenFileResourceTestsWithOpts(t, "data_factory_integration_runtime_managed_test", opts)
}
