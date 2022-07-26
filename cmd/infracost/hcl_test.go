package main_test

import (
	"path"
	"testing"

	"github.com/infracost/infracost/internal/testutil"
)

func TestHCLMultiProjectInfra(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"breakdown", "--config-file", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName(), "infracost.config.yml")},
		nil)
}

func TestHCLMultiWorkspace(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"breakdown", "--config-file", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName(), "infracost.config.yml")},
		nil)
}

func TestHCLMultiVarFiles(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"breakdown",
			"--path", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
			"--terraform-var-file", "var1.tfvars",
			"--terraform-var-file", "var2.tfvars",
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
		nil,
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
		nil,
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

func TestHCLModuleReevaluatedOnInputChange(t *testing.T) {
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

func TestHCLModuleRelativeFilesets(t *testing.T) {
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
		}, &GoldenFileOptions{
			RunTerraformCLI: true,
		},
	)
}

func TestHCLModuleCount(t *testing.T) {
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

func TestHCLModuleForEach(t *testing.T) {
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
