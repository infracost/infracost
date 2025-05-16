package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestMariaDBServer(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	opts := tftest.DefaultGoldenFileOptions()
	// Ignore the CLI because the resource has been removed from the provider in favour of azurerm_mysql_flexible_server
	opts.IgnoreCLI = true

	tftest.GoldenFileResourceTestsWithOpts(t, "mariadb_server_test", opts)
}
