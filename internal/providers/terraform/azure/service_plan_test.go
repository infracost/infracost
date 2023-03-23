package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAzureRMServicePlan(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tftest.GoldenFileResourceTestsWithOpts(t, "service_plan_test", &tftest.GoldenFileOptions{CaptureLogs: true})
}
