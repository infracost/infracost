package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/testutil"
)

func TestDiffHelp(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--help"}, nil)
}

func TestDiffTerraformPlanJSON(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml"}, nil)
}

func TestDiffTerraformDirectory(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--path", "../../examples/terraform", "--terraform-force-cli"}, nil)
}

func TestDiffTerraformShowSkipped(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--path", "./testdata/express_route_gateway_plan.json", "--show-skipped"}, nil)
}

func TestDiffTerraformOutFile(t *testing.T) {
	testdataName := testutil.CalcGoldenFileTestdataDirName()
	goldenFilePath := "./testdata/" + testdataName + "/infracost_output.golden"
	outputPath := filepath.Join(t.TempDir(), "infracost_output.txt")

	GoldenFileCommandTest(t, testdataName, []string{"diff", "--path", "./testdata/example_plan.json", "--out-file", outputPath}, nil)

	actual, err := os.ReadFile(outputPath)
	require.Nil(t, err)
	actual = stripDynamicValues(actual)

	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}

// Need to figure out how to capture synced file before we enable this
// func TestDiffTerraformSyncUsageFile(t *testing.T) {
// 	testdataName := testutil.CalcGoldenFileTestdataDirName()
// 	GoldenFileCommandTest(t, testdataName, []string{"diff", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml", "--sync-usage-file"}, nil)
// }

func TestDiffProjectName(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	GoldenFileCommandTest(t, testName,
		[]string{
			"diff",
			"--config-file", path.Join("./testdata", testName, "infracost-config.yml"),
			"--compare-to", path.Join("./testdata", testName, "prior.json"),
		}, nil)
}

func TestDiffProjectNameNoChange(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	GoldenFileCommandTest(t, testName,
		[]string{
			"diff",
			"--config-file", path.Join("./testdata", testName, "infracost-config.yml"),
			"--compare-to", path.Join("./testdata", testName, "prior.json"),
		}, nil)
}

func TestDiffWithCompareTo(t *testing.T) {
	dir := path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"diff",
			"--path",
			dir,
			"--compare-to",
			path.Join(dir, "prior.json"),
		}, &GoldenFileOptions{
			RunTerraformCLI: true,
		})
}
func TestDiffWithCompareToWithCurrentAndPastProjectError(t *testing.T) {
	dir := path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"diff",
			"--path",
			dir,
			"--compare-to",
			path.Join(dir, "prior.json"),
			"--format", "json",
		}, &GoldenFileOptions{
			IsJSON: true,
		})
}

func TestDiffWithCompareToWithCurrentProjectError(t *testing.T) {

	dir := path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"diff",
			"--path",
			dir,
			"--compare-to",
			path.Join(dir, "prior.json"),
			"--format", "json",
		}, &GoldenFileOptions{
			IsJSON: true,
		})
}

func TestDiffWithCompareToWithPastProjectError(t *testing.T) {
	dir := path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"diff",
			"--path",
			dir,
			"--compare-to",
			path.Join(dir, "prior.json"),
			"--format", "json",
		}, &GoldenFileOptions{
			IsJSON: true,
		})
}

func TestDiffWithCompareToWithPastProjectErrorTerragrunt(t *testing.T) {
	dir := path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())

	//// write a file to the terraform directory which the terragrunt projects will use
	//// this will cause a module source error, meaning an underlying terraform evaluation error
	//// this is to test that terragrunt evaluation properly treats this case as a project error
	//// and not a warning.
	f, err := os.Create(path.Join(dir, "us/modules/example/bad.tf"))
	require.NoError(t, err)
	_, err = f.WriteString("module \"example\" {\n  source = \"doesntexist\"\n}\n")
	assert.NoError(t, err)
	err = f.Close()
	assert.NoError(t, err)

	// now we create the prior.json file by running breakdown
	GetCommandOutput(
		t,
		[]string{
			"breakdown",
			"--config-file",
			path.Join(dir, "infracost.yml"),
			"--format", "json",
			"--output-file",
			path.Join(dir, "prior.json"),
		},
		&GoldenFileOptions{},
	)

	err = os.Remove(path.Join(dir, "us/modules/example/bad.tf"))
	require.NoError(t, err)

	out, err := os.ReadFile(path.Join(dir, "prior.json"))
	require.NoError(t, err)
	out = bytes.ReplaceAll(out, []byte("REPLACED_PROJECT_PATH/"), []byte(""))
	pretty := bytes.NewBuffer(nil)
	err = json.Indent(pretty, out, "", "  ")
	require.NoError(t, err)
	re := regexp.MustCompile(`/.*/doesntexist`)
	out = re.ReplaceAll(pretty.Bytes(), []byte("BAD_MODULE_PATH"))
	err = os.WriteFile(path.Join(dir, "prior.json"), out, 0600)
	require.NoError(t, err)

	// and check the diff is correct
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"diff",
			"--config-file",
			path.Join(dir, "infracost.yml"),
			"--compare-to",
			path.Join(dir, "prior.json"),
		}, nil)
}

func TestDiffWithCompareToFormatJSON(t *testing.T) {
	dir := path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"diff",
			"--path",
			dir,
			"--compare-to",
			path.Join(dir, "prior.json"),
			"--format",
			"json",
		}, &GoldenFileOptions{
			RunTerraformCLI: true,
			IsJSON:          true,
		},
	)
}

func TestDiffWithCompareToNoMetadataFormatJSON(t *testing.T) {
	dir := path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"diff",
			"--path",
			dir,
			"--compare-to",
			path.Join(dir, "prior.json"),
			"--format",
			"json",
		}, &GoldenFileOptions{
			RunTerraformCLI: true,
			IsJSON:          true,
		},
	)
}

func TestDiffWithCompareToPreserveSummary(t *testing.T) {
	dir := path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"diff",
			"--path",
			"./testdata/express_route_gateway_plan.json",
			"--compare-to",
			path.Join(dir, "prior.json"),
			"--show-skipped",
		}, &GoldenFileOptions{
			RunTerraformCLI: true,
		})
}

func TestDiffWithInfracostJSON(t *testing.T) {
	dir := path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"diff",
			"--path",
			path.Join(dir, "current.json"),
			"--compare-to",
			path.Join(dir, "prior.json"),
		}, &GoldenFileOptions{
			RunTerraformCLI: true,
		})
}

func TestDiffWithConfigFileCompareTo(t *testing.T) {
	dir := path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())
	configFile := fmt.Sprintf(`version: 0.1

projects:
  - path: %s
  - path: %s`,
		path.Join(dir, "dev"),
		path.Join(dir, "prod"))

	configFilePath := path.Join(dir, "infracost.yml")
	err := os.WriteFile(configFilePath, []byte(configFile), os.ModePerm) // nolint: gosec
	require.NoError(t, err)

	defer os.Remove(configFilePath)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"diff",
			"--config-file",
			configFilePath,
			"--compare-to",
			path.Join(dir, "prior.json"),
		}, nil)
}

func TestDiffWithConfigFileCompareToDeletedProject(t *testing.T) {
	dir := path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())
	configFile := fmt.Sprintf(`version: 0.1

projects:
  - path: %s`,
		path.Join(dir, "prod"))

	configFilePath := path.Join(dir, "infracost.yml")
	err := os.WriteFile(configFilePath, []byte(configFile), os.ModePerm) // nolint: gosec
	require.NoError(t, err)

	defer os.Remove(configFilePath)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"diff",
			"--config-file",
			configFilePath,
			"--compare-to",
			path.Join(dir, "prior.json"),
		}, nil)
}

func TestDiffCompareToError(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--path", "../../examples/terraform"}, nil)
}

func TestDiffCompareToErrorTerragrunt(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--path", "../../examples/terragrunt"}, nil)
}

func TestDiffTerraformUsageFile(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml"}, nil)
}

func TestDiffTerragrunt(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--path", "../../examples/terragrunt", "--terraform-force-cli"}, nil)
}

func TestDiffTerragruntNested(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"diff", "--path", "../../examples", "--terraform-force-cli"}, nil)
}

func TestDiffTerragruntSyntaxError(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"diff",
			"--compare-to", filepath.Join(dir, "baseline.withouterror.json"),
			"--config-file", filepath.Join(dir, "infracost.config.yml"),
		},
		&GoldenFileOptions{},
	)
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

func TestDiffWithFreeResourcesChecksum(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"diff",
			"--compare-to", filepath.Join(dir, "past.json"),
			"--path", dir,
			"--format", "json",
		}, &GoldenFileOptions{IsJSON: true}, func(ctx *config.RunContext) {
			ctx.Config.TagPoliciesEnabled = true
		})
}

func TestDiffWithPolicyDataUpload(t *testing.T) {
	ts := GraphqlTestServer(map[string]string{
		"policyResourceAllowList": policyResourceAllowlistGraphQLResponse,
		"storePolicyResources":    storePolicyResourcesGraphQLResponse,
	})
	defer ts.Close()

	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testName,
		[]string{
			"diff",
			"--compare-to", filepath.Join(dir, "baseline.json"),
			"--path", dir,
			"--format", "json",
		},
		&GoldenFileOptions{
			CaptureLogs: true,
			IsJSON:      true,
		}, func(ctx *config.RunContext) {
			ctx.Config.PolicyV2APIEndpoint = ts.URL
			ctx.Config.PoliciesEnabled = true
		},
	)
}

func TestDiffPriorEmptyProject(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(), []string{
			"diff",
			"--compare-to",
			path.Join(dir, "base.json"),
			"--path",
			dir,
		}, nil)
}

func TestDiffPriorEmptyProjectJSON(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(), []string{
			"diff",
			"--compare-to",
			path.Join(dir, "base.json"),
			"--path",
			dir,
			"--format", "json",
		}, &GoldenFileOptions{
			IsJSON: true,
		})
}
