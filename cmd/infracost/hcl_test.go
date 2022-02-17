package main_test

import (
	"path"
	"testing"

	"github.com/infracost/infracost/internal/testutil"
)

func TestHCLMultiProjectInfra(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"breakdown", "--config-file", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName(), "infracost.config.yml")},
		&GoldenFileOptions{RunHCL: true})
}
