package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/testutil"
)

func TestBreakdownHelp(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--help"}, nil)
}

func TestBreakdownFormatHTML(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--format", "html", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml"}, nil)
}

func TestBreakdownFormatHTMLProjectName(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--format", "html",
			"--project-name", "my-custom-project-name",
			"--path", "../../examples/terragrunt",
			"--terraform-workspace", "testworkspace",
		}, nil)
}

func TestBreakdownFormatJSON(t *testing.T) {
	opts := DefaultOptions()
	opts.IsJSON = true
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"breakdown", "--format", "json", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml"}, opts)
}

func TestBreakdownFormatJsonWithWarnings(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testName,
		[]string{
			"breakdown",
			"--format", "json",
			"--path", dir,
		},
		DefaultOptions(),
	)
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
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "../../examples/terraform"}, &GoldenFileOptions{RunTerraformCLI: true})
}

func TestBreakdownMultiProjectAutodetect(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", dir,
		}, nil,
	)
}

func TestBreakdownConfigFile(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--config-file", path.Join(dir, "infracost.yml"),
			"--format", "json",
		}, nil,
	)
}

func TestBreakdownMultiProjectSkipPaths(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", dir,
			"--exclude-path", "glob/*/shown",
			"--exclude-path", "ignored",
		}, nil,
	)
}

func TestBreakdownMultiProjectSkipPathsRootLevel(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", dir,
			"--exclude-path", "dev",
			"--exclude-path", "prod",
		}, nil,
	)
}

func TestBreakdownTerraformDirectoryWithDefaultVarFiles(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	t.Run("with terraform plan flags", func(t *testing.T) {
		GoldenFileCommandTest(
			t,
			testName,
			[]string{
				"breakdown",
				"--path", dir,
				"--terraform-plan-flags", "-var-file=input.tfvars -var=block2_ebs_volume_size=2000 -var block2_volume_type=io1",
				"--terraform-force-cli",
			},
			nil,
		)
	})

	t.Run("with hcl var flags", func(t *testing.T) {
		GoldenFileCommandTest(
			t,
			testName,
			[]string{
				"breakdown",
				"--path", dir,
				"--terraform-var-file", "hcl.tfvars",
				"--terraform-var-file", "hcl.tfvars.json",
				"--terraform-var", "block2_ebs_volume_size=2000",
				"--terraform-var", "block2_volume_type=io1",
			},
			nil,
		)
	})

	t.Run("with hcl TF_VAR env variables", func(t *testing.T) {
		GoldenFileCommandTest(
			t,
			testName,
			[]string{
				"breakdown",
				"--path", dir,
				"--terraform-var-file", "input.tfvars",
			},
			&GoldenFileOptions{
				Env: map[string]string{
					"TF_VAR_block2_ebs_volume_size": "2000",
					"TF_VAR_block2_volume_type":     "io1",
				},
			},
		)
	})

	// t.Run("with hcl TF_VAR env variables in config file", func(t *testing.T) {
	//	GoldenFileCommandTest(
	//		t,
	//		testName,
	//		[]string{
	//			"breakdown",
	//			"--config-file", path.Join(dir, "infracost-config.yml"),
	//		},
	//		nil,
	//	)
	// })
}

func TestBreakdownTerraformDirectoryWithRecursiveModules(t *testing.T) {
	dir := path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", dir}, &GoldenFileOptions{RunTerraformCLI: true})
}

func TestBreakdownTerraformFieldsAll(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml", "--fields", "all"}, nil)
}

func TestBreakdownTerraformFieldsInvalid(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml", "--fields", "price,hourlyCost,invalid"}, nil)
}

func TestBreakdownTerraformShowSkipped(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/express_route_gateway_plan.json", "--show-skipped"}, nil)
}

func TestBreakdownTerraformOutFileHTML(t *testing.T) {
	testdataName := testutil.CalcGoldenFileTestdataDirName()
	goldenFilePath := "./testdata/" + testdataName + "/infracost_output.golden"
	outputPath := filepath.Join(t.TempDir(), "infracost_output.html")

	GoldenFileCommandTest(t, testdataName, []string{"breakdown", "--path", "./testdata/example_plan.json", "--format", "html", "--out-file", outputPath}, nil)

	actual, err := ioutil.ReadFile(outputPath)
	require.Nil(t, err)
	actual = stripDynamicValues(actual)

	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}

func TestBreakdownTerraformOutFileJSON(t *testing.T) {
	testdataName := testutil.CalcGoldenFileTestdataDirName()
	goldenFilePath := "./testdata/" + testdataName + "/infracost_output.golden"
	outputPath := filepath.Join(t.TempDir(), "infracost_output.json")

	GoldenFileCommandTest(t, testdataName, []string{"breakdown", "--path", "./testdata/example_plan.json", "--format", "json", "--out-file", outputPath}, nil)

	actual, err := ioutil.ReadFile(outputPath)
	require.Nil(t, err)
	actual = stripDynamicValues(actual)

	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}

func TestBreakdownTerraformOutFileTable(t *testing.T) {
	testdataName := testutil.CalcGoldenFileTestdataDirName()
	goldenFilePath := "./testdata/" + testdataName + "/infracost_output.golden"
	outputPath := filepath.Join(t.TempDir(), "infracost_output.txt")

	GoldenFileCommandTest(t, testdataName, []string{"breakdown", "--path", "./testdata/example_plan.json", "--out-file", outputPath}, nil)

	actual, err := ioutil.ReadFile(outputPath)
	require.Nil(t, err)
	actual = stripDynamicValues(actual)

	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}

func TestBreakdownTerraformSyncUsageFile(t *testing.T) {
	testdataName := testutil.CalcGoldenFileTestdataDirName()
	goldenFilePath := "./testdata/" + testdataName + "/infracost-usage.yml.golden"
	usageFilePath := "./testdata/" + testdataName + "/infracost-usage.yml"

	GoldenFileCommandTest(t, testdataName, []string{"breakdown", "--path", "testdata/breakdown_terraform_sync_usage_file/sync_usage_file.json", "--usage-file", usageFilePath, "--sync-usage-file"}, nil)

	actual, err := ioutil.ReadFile(usageFilePath)
	require.Nil(t, err)
	actual = stripDynamicValues(actual)

	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}

func TestBreakdownTerraformUsageFile(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml"}, nil)
}

func TestBreakdownTerraformUsageFileInvalidKey(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/infracost-usage-invalid-key.yml"}, nil)
}

func TestBreakdownInvalidPath(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "invalid"}, nil)
}

func TestBreakdownPlanError(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "../..//examples/terraform", "--terraform-plan-flags", "-var-file=invalid", "--terraform-force-cli"}, &GoldenFileOptions{CaptureLogs: true})
}

func TestBreakdownTerragrunt(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "../../examples/terragrunt"}, nil)
}

func TestBreakdownTerragruntWithProjectError(t *testing.T) {
	t.Skip()

	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(t,
		testName,
		[]string{
			"breakdown",
			"--path", dir},
		nil)
}

func TestBreakdownTerragruntWithDashboardEnabled(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "../../examples/terragrunt"}, nil, func(c *config.RunContext) {
		c.Config.EnableDashboard = true
		c.Config.EnableCloud = nil
	})
}

func TestBreakdownTerragruntHCLSingle(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "../../examples/terragrunt/prod"}, nil)
}

func TestBreakdownTerragruntHCLMulti(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "../../examples/terragrunt"}, nil)
}

func TestBreakdownTerragruntHCLDepsOutput(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())}, nil)
}

func TestBreakdownTerragruntGetEnv(t *testing.T) {
	os.Setenv("CUSTOM_OS_VAR", "test")
	os.Setenv("CUSTOM_OS_VAR_PROD", "test-prod")
	defer func() {
		os.Unsetenv("CUSTOM_OS_VAR")
		os.Unsetenv("CUSTOM_OS_VAR_PROD")
	}()

}

func TestBreakdownTerragruntHCLDepsOutputMocked(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())}, nil)
}

func TestBreakdownTerragruntSkipPaths(t *testing.T) {
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
			"--exclude-path", "glob/*/ignored",
			"--exclude-path", "ignored",
		},
		nil,
	)
}

func TestBreakdownTerragruntHCLDepsOutputInclude(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()+"/dev")}, nil)
}

func TestBreakdownTerragruntHCLDepsOutputSingleProject(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()+"/dev")}, nil)
}

func TestBreakdownTerragruntHCLMultiNoSource(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/breakdown_terragrunt_hclmulti_no_source/example"}, nil)
}

func TestBreakdownTerragruntNested(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "../../examples"}, nil)
}

func TestBreakdownTerragruntIAMRoles(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())}, nil)
}

func TestBreakdownTerragruntExtraArgs(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())}, nil)
}

func TestInstanceWithAttachmentBeforeDeploy(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/instance_with_attachment_before_deploy.json"}, nil)
}

func TestInstanceWithAttachmentAfterDeploy(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/instance_with_attachment_after_deploy.json"}, nil)
}

func TestBreakdownTerraformWrapper(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/plan_with_terraform_wrapper.json"}, nil)
}

func TestBreakdownWithTarget(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/plan_with_target.json"}, nil)
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

func TestBreakdownInitFlagsError(t *testing.T) {
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path",
			path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
			"--terraform-init-flags",
			"-plugin-dir=does/not/exist",
		},
		nil,
	)
}

func TestBreakdownWithPrivateTerraformRegistryModule(t *testing.T) {
	if os.Getenv("INFRACOST_TERRAFORM_CLOUD_TOKEN") == "" {
		t.Skip("Skipping because INFRACOST_TERRAFORM_CLOUD_TOKEN is not set and external contributors won't have this.")
	}

	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path",
			path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
		},
		nil,
	)
}

func TestBreakdownWithWorkspace(t *testing.T) {
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path",
			path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
			"--terraform-workspace",
			"prod",
		},
		nil,
	)
}

func TestBreakdownWithActualCosts(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := ioutil.ReadAll(r.Body)
		graphqlQuery := string(bodyBytes)

		if strings.Contains(graphqlQuery, "actualCostsList") {
			fmt.Fprintln(w, `[{"data": {"actualCostsList":[{
				"address": "aws_dynamodb_table.usage",
				"resourceId": "arn:aws_dynamodb_table",
				"startAt": "2022-09-15T09:09:09Z",
				"endAt": "2022-09-22T09:09:09Z",
				"costComponents": [{
						"usageType": "someusagetype",
						"description": "$0.005123 per some aws thing",
						"currency": "USD",
						"monthlyCost": "5.123",
						"monthlyQuantity": "1000",
						"price": "0.005123",
						"unit": "GB"
					},
					{
						"usageType": "someusagetype",
						"description": "$0.045 per some other aws thing",
						"currency": "USD",
						"monthlyCost": "45.0",
						"monthlyQuantity": "1000",
						"price": "0.045",
						"unit": "GB"
					}
				]
			},
			{
				"address": "aws_dynamodb_table.usage",
				"resourceId": "arn:another_aws_dynamodb_table",
				"startAt": "2022-08-15T09:09:09Z",
				"endAt": "2022-09-22T09:09:09Z",
				"costComponents": [{
					"usageType": "someusagetype",
					"description": "$0.005123 per some aws thing",
					"currency": "USD",
					"monthlyCost": "5.123",
					"monthlyQuantity": "1000",
					"price": "0.005123",
					"unit": "GB"
				}]
			}
			]}}]`)
		} else if strings.Contains(graphqlQuery, "usageQuantities") {
			keys := []string{
				"monthly_write_request_units",
				"monthly_read_request_units",
				"storage_gb",
				"pitr_backup_storage_gb",
				"on_demand_backup_storage_gb",
				"monthly_data_restored_gb",
				"monthly_streams_read_request_units",
			}

			keyRows := make([]string, len(keys))
			for i, k := range keys {
				keyRows[i] = fmt.Sprintf(`{"address": "aws_dynamodb_table.usage", "usageKey": "%s", "monthlyQuantity": "%d"}`, k, 100000+i)
			}
			fmt.Fprintf(w, `[{"data": {"usageQuantities":[%s]}}]`, strings.Join(keyRows, ","))
		}
	}))
	defer ts.Close()

	GoldenFileCommandTest(t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{"breakdown", "--path", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())},
		&GoldenFileOptions{CaptureLogs: true},
		func(c *config.RunContext) {
			c.Config.UsageAPIEndpoint = ts.URL
			c.Config.UsageActualCosts = true
		},
	)
}

func TestBreakdownWithNestedForeach(t *testing.T) {
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path",
			path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
		},
		nil,
	)
}

func TestBreakdownWithDependsUponModule(t *testing.T) {
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path",
			path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
		},
		nil,
	)
}

func TestBreakdownWithOptionalVariables(t *testing.T) {
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path",
			path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
		},
		nil,
	)
}

func TestBreakdownWithDeepMergeModule(t *testing.T) {
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path",
			path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
		},
		nil,
	)
}

func TestBreakdownWithMultipleProviders(t *testing.T) {
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path",
			path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
		},
		nil,
	)
}

func TestBreakdownMultiProjectWithError(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", dir,
		}, &GoldenFileOptions{CaptureLogs: true},
	)
}

func TestBreakdownMultiProjectWithErrorOutputJSON(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testName,
		[]string{
			"breakdown",
			"--format", "json",
			"--path",
			dir,
			"--format",
			"json",
		}, &GoldenFileOptions{CaptureLogs: true},
	)
}

func TestBreakdownMultiProjectWithAllErrors(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", dir,
		}, &GoldenFileOptions{CaptureLogs: true},
	)
}
