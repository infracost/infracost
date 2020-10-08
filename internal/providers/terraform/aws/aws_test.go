package aws_test

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Short() {
		// Ensure plugins are installed and cached
		err := tftest.InstallPlugins()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	code := m.Run()
	os.Exit(code)
}
