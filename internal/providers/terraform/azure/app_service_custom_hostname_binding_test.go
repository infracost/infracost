package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAzureRMAppServiceCustomHostnameBinding(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	opts := tftest.DefaultGoldenFileOptions()
	opts.RequiresInit = true
	tftest.GoldenFileResourceTestsWithOpts(t, "app_service_custom_hostname_binding_test", opts)
}
