package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAzureRMWindowsVirtualMachineGoldenFile(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	t.Run("base price", func(t *testing.T) {
		tftest.GoldenFileResourceTestsWithOpts(t, "windows_virtual_machine_test", &tftest.GoldenFileOptions{
			IgnoreCLI: true,
		})
	})
	t.Run("dev/test price", func(t *testing.T) {
		tftest.GoldenFileResourceTestsWithOpts(t, "windows_virtual_machine_test", &tftest.GoldenFileOptions{
			IgnoreCLI:        true,
			GoldenFileSuffix: "dev_test_price",
		}, func(ctx *config.RunContext) {
			ctx.Config.Projects[0].Metadata = map[string]string{
				"isProduction": "false",
			}
		})
	})
}
