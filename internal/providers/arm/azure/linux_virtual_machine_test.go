package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/arm/armtest"
)

func TestAzureRMLinuxVirtualMachineGoldenFile(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	armtest.GoldenFileResourceTests(t, "linux_virtual_machine_test")
}
