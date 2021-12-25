package main_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/infracost/infracost/internal/testutil"
)

func TestDiffHelp(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--help"}, nil)
}

func TestDiffTerraformPlanJSON(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml"}, nil)
}

func TestDiffTerraformDirectory(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--path", "../../examples/terraform"}, nil)
}

func TestDiffTerraformShowSkipped(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--path", "./testdata/express_route_gateway_plan.json", "--show-skipped"}, nil)
}

func TestDiffTerraformOutFile(t *testing.T) {
	testdataName := testutil.CalcGoldenFileTestdataDirName()
	goldenFilePath := "./testdata/" + testdataName + "/infracost_output.golden"
	outputPath := filepath.Join(t.TempDir(), "infracost_output.txt")

	GoldenFileCommandTest(t, testdataName, []string{"diff", "--path", "./testdata/example_plan.json", "--out-file", outputPath}, nil)

	actual, err := ioutil.ReadFile(outputPath)
	require.Nil(t, err)
	actual = stripDynamicValues(actual)

	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}

// Need to figure out how to capture synced file before we enable this
// func TestDiffTerraformSyncUsageFile(t *testing.T) {
// 	testdataName := testutil.CalcGoldenFileTestdataDirName()
// 	GoldenFileCommandTest(t, testdataName, []string{"diff", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml", "--sync-usage-file"}, nil)
// }

func TestDiffTerraformUsageFile(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml"}, nil)
}

func TestDiffTerragrunt(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--path", "../../examples/terragrunt"}, nil)
}

func TestDiffTerragruntNested(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--path", "../../examples"}, nil)
}

func TestDiffWithTarget(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--path", "./testdata/plan_with_target.json"}, nil)
}

func TestDiffTerraform_v0_12(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--path", "./testdata/terraform_v0.12_plan.json"}, nil)
}

func TestDiffTerraform_v0_14(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--path", "./testdata/terraform_v0.14_plan.json"}, nil)
}
