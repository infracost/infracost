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

func TestHCLMultiWorkspace(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"breakdown", "--config-file", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName(), "infracost.config.yml")},
		&GoldenFileOptions{RunHCL: true})
}

func TestHCLMultiVarFiles(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"breakdown",
			"--path", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
			"--terraform-var-file", "var1.tfvars",
			"--terraform-var-file", "var2.tfvars",
			"--terraform-plan-flags=-var-file=./var1.tfvars -var-file=./var2.tfvars",
		},
		&GoldenFileOptions{RunHCL: true})
}
