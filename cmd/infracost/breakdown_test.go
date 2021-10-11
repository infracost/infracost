package main_test

import (
	"testing"

	"github.com/infracost/infracost/internal/testutil"
)

func TestBreakdownHelp(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--help"}, nil)
}

func TestBreakdownFormatHTML(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--format", "html", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml"}, nil)
}

func TestBreakdownFormatJSON(t *testing.T) {
	opts := DefaultOptions()
	opts.IsJSON = true
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--format", "json", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml"}, opts)
}

func TestBreakdownFormatJSONShowSkipped(t *testing.T) {
	opts := DefaultOptions()
	opts.IsJSON = true
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--format", "json", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml", "--show-skipped"}, opts)
}

func TestBreakdownFormatTable(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--format", "table", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml"}, nil)
}

func TestBreakdownTerraformPlanJSON(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml"}, nil)
}

func TestBreakdownTerraformDirectory(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "../../examples/terraform"}, nil)
}

func TestBreakdownTerraformFieldsAll(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml", "--fields", "all"}, nil)
}

func TestBreakdownTerraformFieldsInvalid(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml", "--fields", "price,hourlyCost,invalid"}, nil)
}

func TestBreakdownTerraformShowSkipped(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/azure_firewall_plan.json", "--show-skipped"}, nil)
}

// Need to figure out how to capture synced file before we enable this
// func TestBreakdownTerraformSyncUsageFile(t *testing.T) {
// 	testdataName := testutil.CalcGoldenFileTestdataDirName()
// 	GoldenFileCommandTest(t, testdataName, []string{"breakdown", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml", "--sync-usage-file" }, nil)
// }

func TestBreakdownTerraformUsageFile(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml"}, nil)
}

func TestBreakdownTerragrunt(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "../../examples/terragrunt"}, nil)
}

func TestBreakdownTerragruntNested(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "../../examples"}, nil)
}

func TestBreakdownTerraform_v0_12(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/terraform_v0.12_plan.json"}, nil)
}

func TestBreakdownTerraformUseState_v0_12(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/terraform_v0.12_state.json", "--terraform-use-state"}, nil)
}

func TestBreakdownTerraform_v0_14(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/terraform_v0.14_plan.json"}, nil)
}

func TestBreakdownTerraformUseState_v0_14(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/terraform_v0.14_state.json", "--terraform-use-state"}, nil)
}
