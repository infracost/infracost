package main_test

import (
	"github.com/infracost/infracost/internal/testutil"
	"testing"
)

func TestDiffHelp(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--help"}, nil)
}

func TestDiffTerraformPlanJSON(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--path", "../../examples/terraform/plan.json"}, nil)
}

func TestDiffTerraformDirectory(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--path", "../../examples/terraform"}, nil)
}

func TestDiffTerraformShowSkipped(t *testing.T) {
	testdataName := testutil.CalcGoldenFileTestdataDirName()
	GoldenFileCommandTest(t, testdataName, []string{"diff", "--path", "./testdata/" + testdataName + "/plan.json", "--show-skipped"}, nil)
}

// Need to figure out how to capture synced file before we enable this
// func TestDiffTerraformSyncUsageFile(t *testing.T) {
// 	testdataName := testutil.CalcGoldenFileTestdataDirName()
// 	GoldenFileCommandTest(t, testdataName, []string{"diff", "--path", "../../examples/terraform/plan.json", "--usage-file", "./testdata/" + testdataName + "/infracost-usage.yml", "--sync-usage-file"}, nil)
// }

func TestDiffTerraformUsageFile(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--path", "../../examples/terraform/plan.json", "--usage-file", "../../examples/terraform/infracost-usage.yml"}, nil)
}
