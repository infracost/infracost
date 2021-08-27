package main_test

import (
	"github.com/infracost/infracost/internal/testutil"
	"testing"
)

func TestBreakdownHelp(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--help"}, nil)
}

func TestBreakdownFormatHTML(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--format", "html", "--path", "../../examples/terraform/plan.json"}, nil)
}

func TestBreakdownFormatJSON(t *testing.T) {
	opts := DefaultOptions()
	opts.IsJSON = true
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--format", "json", "--path", "../../examples/terraform/plan.json"}, opts)
}

func TestBreakdownFormatTable(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--format", "table", "--path", "../../examples/terraform/plan.json"}, nil)
}

func TestBreakdownTerraformPlanJSON(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "../../examples/terraform/plan.json"}, nil)
}

func TestBreakdownTerraformDirectory(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "../../examples/terraform"}, nil)
}

func TestBreakdownTerraformFieldsAll(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "../../examples/terraform/plan.json", "--fields", "all"}, nil)
}

func TestBreakdownTerraformShowSkipped(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/azure_firewall_plan.json", "--show-skipped"}, nil)
}

// Need to figure out how to capture synced file before we enable this
// func TestBreakdownTerraformSyncUsageFile(t *testing.T) {
// 	testdataName := testutil.CalcGoldenFileTestdataDirName()
// 	GoldenFileCommandTest(t, testdataName, []string{"breakdown", "--path", "../../examples/terraform/plan.json", "--usage-file", "./testdata/" + testdataName + "/infracost-usage.yml", "--sync-usage-file"}, nil)
// }

func TestBreakdownTerraformUsageFile(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "../../examples/terraform/plan.json", "--usage-file", "../../examples/terraform/infracost-usage.yml"}, nil)
}
