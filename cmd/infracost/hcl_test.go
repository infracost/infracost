package main_test

import (
	"path"
	"testing"

	"github.com/infracost/infracost/internal/testutil"
)

func TestHCLMultiProjectInfra(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	GoldenFileCommandTest(t, testName,
		[]string{"breakdown", "--log-level", "info", "--config-file", path.Join("./testdata", testName, "infracost.config.yml")}, nil)
}

func TestHCLMultiWorkspace(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"breakdown", "--config-file", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName(), "infracost.config.yml")}, nil)
}

func TestHCLMultiVarFiles(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"breakdown",
			"--path", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
			"--terraform-var-file", "var1.tfvars",
			"--terraform-var-file", "var2.tfvars",
			"--terraform-plan-flags=-var-file=./var1.tfvars -var-file=./var2.tfvars",
		},
		nil,
	)
}

func TestHCLProviderAlias(t *testing.T) {
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
		},
		&GoldenFileOptions{RunTerraformDir: true},
	)
}

func TestHCLModuleOutputCounts(t *testing.T) {
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
		},
		&GoldenFileOptions{RunTerraformDir: true},
	)
}

func TestHCLModuleOutputCountsNested(t *testing.T) {
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
		},
		nil,
	)
}
