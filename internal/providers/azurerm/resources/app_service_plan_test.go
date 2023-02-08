package resources_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/azurerm/armtest"
)

func TestAzureRMAppServicePlan(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	armtest.GoldenFileResourceTests(t, "app_service_plan_test")
}
