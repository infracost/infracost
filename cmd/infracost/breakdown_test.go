package main_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

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
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--format", "json", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml"}, opts)
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
		&GoldenFileOptions{
			CaptureLogs: true,
			IsJSON:      true,
		},
	)
}

func TestBreakdownFormatJsonWithTags(t *testing.T) {
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
		&GoldenFileOptions{
			CaptureLogs: true,
			IsJSON:      true,
		}, func(ctx *config.RunContext) {
			ctx.Config.TagPoliciesEnabled = true
		},
	)
}

func TestBreakdownFormatJsonWithTagsVersionLessThan5point39(t *testing.T) {
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
		&GoldenFileOptions{
			CaptureLogs: true,
			IsJSON:      true,
		}, func(ctx *config.RunContext) {
			ctx.Config.TagPoliciesEnabled = true
		},
	)
}

func TestBreakdownFormatJsonWithTagsAftModule(t *testing.T) {
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
		&GoldenFileOptions{
			CaptureLogs: true,
			IsJSON:      true,
			JSONInclude: regexp.MustCompile("^(defaultTags|tags|name)$"),
			JSONExclude: regexp.MustCompile("^(costComponents|pastBreakdown)$"),
			RegexFilter: regexp.MustCompile("(tags-infracost-mock-44956be29f34|bucket-infracost-mock-44956be29f34|name-infracost-mock-44956be29f34)"),
		}, func(ctx *config.RunContext) {
			ctx.Config.TagPoliciesEnabled = true
		},
	)
}

func TestBreakdownFormatJsonWithTagsAliasedProvider(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)

	GoldenFileCommandTest(
		t,
		testName,
		[]string{
			"breakdown",
			"--format", "json",
			"--path", path.Join(dir, "project"),
		},
		&GoldenFileOptions{
			CaptureLogs: true,
			IsJSON:      true,
			JSONInclude: regexp.MustCompile("^(defaultTags|tags|name)$"),
			JSONExclude: regexp.MustCompile("^(costComponents|pastBreakdown)$"),
		}, func(ctx *config.RunContext) {
			ctx.Config.TagPoliciesEnabled = true
		},
	)
}

func TestBreakdownFormatJsonWithTagsAzure(t *testing.T) {
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
		&GoldenFileOptions{
			CaptureLogs: true,
			IsJSON:      true,
		},
	)
}

func TestBreakdownFormatJsonWithTagsGoogle(t *testing.T) {
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
		&GoldenFileOptions{
			CaptureLogs: true,
			IsJSON:      true,
		},
		func(ctx *config.RunContext) {
			ctx.Config.TagPoliciesEnabled = true
		},
	)
}

func TestBreakdownFormatJsonPropagateDefaultsToVolumeTags(t *testing.T) {
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
		&GoldenFileOptions{
			CaptureLogs: true,
			IsJSON:      true,
			JSONInclude: regexp.MustCompile("^(tags|name)$"),
			JSONExclude: regexp.MustCompile("^(costComponents|metadata|pastBreakdown|subresources)$"),
		}, func(ctx *config.RunContext) {
			ctx.Config.TagPoliciesEnabled = true
		},
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
		},
		&GoldenFileOptions{
			IsJSON: true,
		},
	)
}

func TestBreakdownConfigFileWithUsageFile(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--config-file", path.Join(dir, "infracost.yml"),
			"--usage-file", path.Join(dir, "infracost-usage.yml"),
		},
		&GoldenFileOptions{
			IsJSON: true,
		},
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

func TestBreakdownArmTemplateConfigFile(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--config-file", path.Join(dir, "infracost.yml"),
		},
		&GoldenFileOptions{
			IsJSON: true,
		},
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
			}, &GoldenFileOptions{IgnoreNonGraph: true},
		)
	})

	t.Run("with hcl var flags", func(t *testing.T) {
		abs, _ := filepath.Abs(path.Join(dir, "abs.tfvars"))
		GoldenFileCommandTest(
			t,
			testName,
			[]string{
				"breakdown",
				"--path", dir,
				"--terraform-var-file", "hcl.tfvars",
				"--terraform-var-file", abs,
				"--terraform-var-file", "hcl.tfvars.json",
				"--terraform-var", "block2_ebs_volume_size=2000",
				"--terraform-var", "block2_volume_type=io1",
			},
			&GoldenFileOptions{IgnoreNonGraph: true},
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

func TestBreakdownTerraformProvidedDefaultEnvs(t *testing.T) {
	dir := path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--config-file", path.Join(dir, "infracost.yml")}, nil)
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

	actual, err := os.ReadFile(outputPath)
	require.Nil(t, err)
	actual = stripDynamicValues(actual)

	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}

func TestBreakdownTerraformOutFileJSON(t *testing.T) {
	testdataName := testutil.CalcGoldenFileTestdataDirName()
	goldenFilePath := "./testdata/" + testdataName + "/infracost_output.golden"
	outputPath := filepath.Join(t.TempDir(), "infracost_output.json")

	GoldenFileCommandTest(t, testdataName, []string{"breakdown", "--path", "./testdata/example_plan.json", "--format", "json", "--out-file", outputPath}, nil)

	// prettify the output
	file, err := os.ReadFile(outputPath)
	if err != nil {
		t.Error(err)
		return
	}

	data := map[string]interface{}{}
	err = json.Unmarshal(file, &data)
	if err != nil {
		t.Error(err)
		return
	}

	pretty, _ := json.MarshalIndent(data, "", "  ")
	err = os.WriteFile(outputPath, pretty, 0600)
	if err != nil {
		t.Error(err)
		return
	}

	actual, err := os.ReadFile(outputPath)
	require.Nil(t, err)
	actual = stripDynamicValues(actual)

	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}

func TestBreakdownTerraformOutFileTable(t *testing.T) {
	testdataName := testutil.CalcGoldenFileTestdataDirName()
	goldenFilePath := "./testdata/" + testdataName + "/infracost_output.golden"
	outputPath := filepath.Join(t.TempDir(), "infracost_output.txt")

	GoldenFileCommandTest(t, testdataName, []string{"breakdown", "--path", "./testdata/example_plan.json", "--out-file", outputPath}, nil)

	actual, err := os.ReadFile(outputPath)
	require.Nil(t, err)
	actual = stripDynamicValues(actual)

	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}

func TestBreakdownTerraformSyncUsageFile(t *testing.T) {
	testdataName := testutil.CalcGoldenFileTestdataDirName()
	goldenFilePath := "./testdata/" + testdataName + "/infracost-usage.yml.golden"
	usageFilePath := "./testdata/" + testdataName + "/infracost-usage.yml"

	GoldenFileCommandTest(t, testdataName, []string{"breakdown", "--path", "testdata/breakdown_terraform_sync_usage_file/sync_usage_file.json", "--usage-file", usageFilePath, "--sync-usage-file"}, nil)

	actual, err := os.ReadFile(usageFilePath)
	require.Nil(t, err)
	actual = stripDynamicValues(actual)

	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}

func TestBreakdownTerraformUsageFile(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml"}, nil)
}

func TestBreakdownTerraformUsageFileWildcardModule(t *testing.T) {
	name := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())
	GoldenFileCommandTest(
		t,
		name,
		[]string{
			"breakdown",
			"--path", dir,
			"--usage-file", filepath.Join(dir, "infracost-usage.yml"),
		},
		nil,
	)
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

func TestBreakdownTerragruntWithRemoteSource(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	wd, err := os.Getwd()
	require.NoError(t, err)
	cacheDir := filepath.Join(wd, ".infracost")
	require.NoError(t, os.RemoveAll(cacheDir))

	GoldenFileCommandTest(
		t,
		testName,
		[]string{
			"breakdown",
			"--config-file", filepath.Join(dir, "infracost.yml"),
		},
		&GoldenFileOptions{
			IgnoreNonGraph: true,
		},
	)

	dirs, err := getGitBranchesInDirs(filepath.Join(cacheDir, ".terragrunt-cache"))
	require.NoError(t, err)

	// check that there are 5 directories in the download directory as 3 of the 7 projects use the same ref,
	// but one of these has a generate block.
	require.Len(t, dirs, 5)

	assert.Equal(t, "1f94a2fd07b3e29deea4706b5d2fdc68c1d02aad", dirs["nY--fMGWRYoen8N6OqfdCB8CBnc"])
	assert.Equal(t, "b6fa04f65bdcb26869215fb840f5ee088a096bc8", dirs["F_iCrGgnzJEf5w4HUBrbeCRMQo0"])
	assert.Equal(t, "b6fa04f65bdcb26869215fb840f5ee088a096bc8", dirs["F_iCrGgnzJEf5w4HUBrbeCRMQo0-lQzAPeXdDdx4LOk968nZQH7SIN0"])
	assert.Equal(t, "74725d6e91aa91d7283642b7ed3316d12f271212", dirs["KaCyCQXaN6-S634qsDqQ2wwYENc"])
	assert.Equal(t, "master", dirs["vo8pQqWUeCu_1_TBy7LGvx51SW0"])
}

func getGitRef(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	branch := strings.TrimRight(string(output), "\n")

	if branch == "HEAD" {
		// in detached head state, get the commit hash
		cmd := exec.Command("git", "rev-parse", "HEAD")
		cmd.Dir = dir
		output, err = cmd.Output()
		if err != nil {
			return "", err
		}
		branch = strings.TrimRight(string(output), "\n")
	}

	return branch, nil
}

func getGitBranchesInDirs(root string) (map[string]string, error) {
	var dirs = make(map[string]string)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && filepath.Base(path) == ".git" {
			dir := filepath.Dir(path)
			branch, err := getGitRef(dir)
			if err != nil {
				return err
			}

			dirs[filepath.Base(dir)] = branch
			return filepath.SkipDir
		}
		return nil
	})

	return dirs, err
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

func TestBreakdownTerragruntWithMockedFunctions(t *testing.T) {
	GoldenFileCommandTest(t,
		testutil.CalcGoldenFileTestdataDirName(), []string{
			"breakdown",
			"--path", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
		}, &GoldenFileOptions{
			RunTerraformCLI: false,
		})
}

func TestBreakdownTerragruntHCLSingle(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "../../examples/terragrunt/prod"}, nil)
}

func TestBreakdownTerragruntHCLMulti(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "../../examples/terragrunt"}, nil)
}

func TestBreakdownTerragruntHCLDepsOutput(t *testing.T) {
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{"breakdown", "--path",
			path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())},
		&GoldenFileOptions{CaptureLogs: true},
	)
}

func TestBreakdownTerragruntIncludeDeps(t *testing.T) {
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{"breakdown", "--path",
			path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())},
		&GoldenFileOptions{CaptureLogs: true},
	)
}

func TestBreakdownTerragruntHCLModuleOutputForEach(t *testing.T) {
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{"breakdown", "--path",
			path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())},
		&GoldenFileOptions{CaptureLogs: true},
	)
}

func TestBreakdownTerragruntDiffProjectError(t *testing.T) {
	projectPath := path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())

	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"diff",
			"--path", projectPath,
			"--compare-to", path.Join(projectPath, "prior.json"),
		},
		&GoldenFileOptions{CaptureLogs: true},
	)
}

func TestBreakdownTerragruntGetEnv(t *testing.T) {
	os.Setenv("CUSTOM_OS_VAR", "test")
	os.Setenv("CUSTOM_OS_VAR_PROD", "test-prod")
	defer func() {
		os.Unsetenv("CUSTOM_OS_VAR")
		os.Unsetenv("CUSTOM_OS_VAR_PROD")
	}()

	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())}, nil)
}

func TestBreakdownTerragruntGetEnvConfigFile(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--config-file", path.Join(dir, "infracost.yml"),
		},
		nil,
	)
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

func TestBreakdownTerragruntWithParentInclude(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())}, nil)
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

func TestBreakdownTerragruntSourceMap(t *testing.T) {
	t.Setenv("INFRACOST_TERRAFORM_SOURCE_MAP", "git::https://github.com/someorg/terraform_modules.git=../../../../examples")

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

func TestBreakdownTerraformTFJSON(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())}, nil)
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

func TestBreakdownWithPrivateSshModulePopulatesErrors(t *testing.T) {
	output := GetCommandOutput(
		t,
		[]string{
			"breakdown",
			"--path",
			path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
			"--format",
			"json",
		},
		&GoldenFileOptions{
			Env: map[string]string{
				"GIT_TERMINAL_PROMPT": "0",
			},
		},
	)

	res := gjson.ParseBytes(output)
	errs := res.Get("projects.0.metadata.errors").Array()
	require.Len(t, errs, 1)

	code := errs[0].Get("code").Int()
	msg := errs[0].Get("message").String()
	data := errs[0].Get("data").Map()

	assert.Equal(t, 201, int(code))
	assert.Contains(t, msg, "Failed to download module \"git@github.com:someorg/terraform_modules.git\"")
	assert.Equal(t, data["source"].String(), "ssh")
	assert.Equal(t, data["moduleSource"].String(), "git@github.com:someorg/terraform_modules.git")
}

func TestBreakdownWithPrivateHttpsModulePopulatesErrors(t *testing.T) {
	output := GetCommandOutput(
		t,
		[]string{
			"breakdown",
			"--path",
			path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
			"--format",
			"json",
		},
		nil,
	)

	res := gjson.ParseBytes(output)
	errs := res.Get("projects.0.metadata.errors").Array()
	require.Len(t, errs, 1)

	code := errs[0].Get("code").Int()
	msg := errs[0].Get("message").String()
	data := errs[0].Get("data").Map()

	assert.Equal(t, 201, int(code))
	assert.Contains(t, msg, "Failed to download module \"github.com/someorg/terraform_modules.git\"")
	assert.Equal(t, data["source"].String(), "https")
	assert.Equal(t, data["moduleSource"].String(), "github.com/someorg/terraform_modules.git")
}

func TestBreakdownWithPrivateTerraformRegistryModulePopulatesErrors(t *testing.T) {
	t.Setenv("INFRACOST_TERRAFORM_CLOUD_TOKEN", "badkey")

	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path",
			path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
			"--format",
			"json",
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
		bodyBytes, _ := io.ReadAll(r.Body)
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

func TestBreakdownWithActualCostsPitrDisabled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
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

func TestBreakdownWithDefaultTags(t *testing.T) {
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
			"--format", "json",
		},
		&GoldenFileOptions{
			IsJSON: true,
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

func TestBreakdownWithComplexConfigFileVars(t *testing.T) {
	dir := path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())

	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--config-file",
			path.Join(dir, "infracost.yml"),
		},
		nil,
	)
}

func TestBreakdownWithComplexVarFlags(t *testing.T) {
	dir := path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())

	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path",
			dir,
			"--terraform-var",
			"instance_config={\"instance_type\":\"t2.micro\",\"storage\":20}",
			"--terraform-var",
			"lambda_configs=[{memory_size = 128},{ memory_size = 256}]",
			"--usage-file",
			path.Join(dir, "infracost-usage.yml"),
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

func TestBreakdownWithNestedProviderAliases(t *testing.T) {
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

func TestBreakdownWithProvidersDependingOnData(t *testing.T) {
	// This test doesn't pass for the non-graph evaluator
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path",
			path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName()),
		},
		&GoldenFileOptions{IgnoreNonGraph: true},
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
		}, &GoldenFileOptions{
			CaptureLogs: true,
			IsJSON:      true,
		},
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

func TestBreakdownWithLocalPathDataBlock(t *testing.T) {
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

func TestBreakdownAutoWithMultiVarfileProjects(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", dir,
		}, nil)
}

func TestBreakdownWithFreeResourcesChecksum(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", dir,
			"--format", "json",
		}, &GoldenFileOptions{IsJSON: true}, func(ctx *config.RunContext) {
			ctx.Config.TagPoliciesEnabled = true
		})
}

func TestBreakdownWithDataBlocksInSubmod(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", dir,
		}, nil)
}

func TestBreakdownWithPolicyDataUploadHCL(t *testing.T) {
	sep := []byte("===\n")
	ts, uploadWriter := GraphqlTestServerWithWriter(map[string]string{
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
			"breakdown",
			"--format", "json",
			"--path", dir,
		},
		&GoldenFileOptions{
			CaptureLogs: true,
			IsJSON:      true,
		}, func(ctx *config.RunContext) {
			ctx.Config.PolicyV2APIEndpoint = ts.URL
			ctx.Config.PoliciesEnabled = true
			uploadWriter.Write(sep)
		},
	)

	pp := strings.Split(uploadWriter.String(), string(sep))
	for _, p := range pp[1:] {
		testutil.AssertGoldenFile(t, path.Join("./testdata", testName, testName+"-upload.golden"), []byte(p))
	}
}

func TestBreakdownWithMockedMerge(t *testing.T) {
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

func TestBreakdownWithDynamicIterator(t *testing.T) {
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

func TestBreakdownWithPolicyDataUploadPlanJson(t *testing.T) {
	sep := []byte("===\n")
	ts, uploadWriter := GraphqlTestServerWithWriter(map[string]string{
		"policyResourceAllowList": policyResourceAllowlistGraphQLResponse,
		"storePolicyResources":    storePolicyResourcesGraphQLResponse,
	})
	defer ts.Close()

	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName, "plan.json")
	GoldenFileCommandTest(
		t,
		testName,
		[]string{
			"breakdown",
			"--format", "json",
			"--path", dir,
		},
		&GoldenFileOptions{
			CaptureLogs: true,
			IsJSON:      true,
		}, func(ctx *config.RunContext) {
			ctx.Config.PolicyV2APIEndpoint = ts.URL
			ctx.Config.PoliciesEnabled = true
			uploadWriter.Write(sep)
		},
	)

	pp := strings.Split(uploadWriter.String(), string(sep))
	for _, p := range pp[1:] {
		testutil.AssertGoldenFile(t, path.Join("./testdata", testName, testName+"-upload.golden"), []byte(p))
	}
}

func TestBreakdownWithPolicyDataUploadTerragrunt(t *testing.T) {
	sep := []byte("===\n")
	ts, uploadWriter := GraphqlTestServerWithWriter(map[string]string{
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
			"breakdown",
			"--format", "json",
			"--path", dir,
		},
		&GoldenFileOptions{
			CaptureLogs: true,
			IsJSON:      true,
		}, func(ctx *config.RunContext) {
			ctx.Config.PolicyV2APIEndpoint = ts.URL
			ctx.Config.PoliciesEnabled = true
			uploadWriter.Write(sep)
		},
	)

	pp := strings.Split(uploadWriter.String(), string(sep))
	for _, p := range pp[1:] {
		testutil.AssertGoldenFile(t, path.Join("./testdata", testName, testName+"-upload.golden"), stripDynamicValues([]byte(p)))
	}
}

func TestBreakdownConfigFileWithSkipAutoDetect(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--config-file", path.Join(dir, "infracost.yml"),
		},
		&GoldenFileOptions{
			IsJSON: true,
		},
	)
}

func TestBreakdownInvalidAPIKey(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", dir,
		},
		nil,
		func(ctx *config.RunContext) {
			ctx.Config.APIKey = "BAD_KEY"
			ctx.Config.Credentials.APIKey = "BAD_KEY"
		},
	)
}

func TestBreakdownEmptyAPIKey(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", dir,
		},
		nil,
		func(ctx *config.RunContext) {
			ctx.Config.APIKey = ""
			ctx.Config.Credentials.APIKey = ""
		},
	)
}

func TestBreakdownSkipAutodetectionIfTerraformVarFilePassed(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", dir,
			"--terraform-var-file",
			"prod.tfvars",
		},
		nil,
	)
}

func TestBreakdownTerragruntFileFuncs(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") == "" {
		t.Skip("skipping as this test is only designed for GitHub Actions")
	}

	t.Setenv("INFRACOST_CI_PLATFORM", "github_app")
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", dir,
		},
		&GoldenFileOptions{IgnoreLogs: true, IgnoreNonGraph: true},
	)
}

func TestBreakdownNoPricesWarnings(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", dir,
		},
		nil,
	)
}

func TestBreakdownTerraformFileFuncs(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") == "" {
		t.Skip("skipping as this test is only designed for GitHub Actions")
	}

	t.Setenv("INFRACOST_CI_PLATFORM", "github_app")

	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", dir,
		},
		&GoldenFileOptions{IgnoreLogs: true, IgnoreNonGraph: true},
	)
}

func TestBreakdownAutodetectionOutput(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", dir,
		},
		&GoldenFileOptions{LogLevel: strPtr("info")},
	)
}

func TestBreakdownAutodetectionConfigFileOutput(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--config-file", filepath.Join(dir, "infracost.yml"),
			"--log-level", "info",
		},
		&GoldenFileOptions{LogLevel: strPtr("info"), IgnoreNonGraph: true},
	)
}

func TestBreakdownTerragruntAutodetectionOutput(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", dir,
		},
		&GoldenFileOptions{LogLevel: strPtr("info")},
	)
}

func TestBreakdownTerragruntAutodetectionConfigFileOutput(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--config-file", filepath.Join(dir, "infracost.yml"),
			"--log-level", "info",
		},
		&GoldenFileOptions{LogLevel: strPtr("info"), IgnoreNonGraph: true},
	)
}

func TestBreakdownTerragruntPartialInputs(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	dir := path.Join("./testdata", testName)
	GoldenFileCommandTest(
		t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", dir,
			"--log-level", "info",
		},
		&GoldenFileOptions{LogLevel: strPtr("info"), IgnoreNonGraph: true},
	)
}

func strPtr(s string) *string {
	return &s
}
