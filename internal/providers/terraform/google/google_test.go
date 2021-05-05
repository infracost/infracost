package google_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

// Use a single tmp dir for all tests.  This means we only wait for terraform init once.
var tmpDir string

func TestMain(m *testing.M) {
	var err error
	tmpDir, err = ioutil.TempDir("", "tmp_google_test_*")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir) // clean up

	tftest.EnsurePluginsInstalled(tmpDir)
	code := m.Run()
	os.Exit(code)
}
