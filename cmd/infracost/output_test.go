package main_test

import (
	"io/ioutil"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/infracost/infracost/internal/testutil"
)

func TestOutputHelp(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--help"}, nil)
}

func TestOutputFormatHTML(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--format", "html", "--path", "./testdata/example_out.json", "--path", "./testdata/azure_firewall_out.json"}, nil)
}

func TestOutputFormatJSON(t *testing.T) {
	opts := DefaultOptions()
	opts.IsJSON = true
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--format", "json", "--path", "./testdata/example_out.json", "--path", "./testdata/azure_firewall_out.json"}, opts)
}

func TestOutputFormatBitbucketCommentWithProjectNames(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	GoldenFileCommandTest(t, testName,
		[]string{
			"output",
			"--format", "bitbucket-comment",
			"--path", path.Join("./testdata", testName, "infracost.json"),
		}, nil)
}

func TestOutputFormatBitbucketCommentSummary(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--format", "bitbucket-comment-summary", "--path", "./testdata/terraform_v0.14_nochange_breakdown.json"}, nil)
}

func TestOutputFormatGitHubComment(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--format", "github-comment", "--path", "./testdata/example_out.json", "--path", "./testdata/terraform_v0.14_breakdown.json", "--path", "./testdata/terraform_v0.14_nochange_breakdown.json"}, nil)
}

func TestOutputFormatGitHubCommentWithProjectNames(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	GoldenFileCommandTest(t, testName,
		[]string{
			"output",
			"--format", "github-comment",
			"--path", path.Join("./testdata", testName, "infracost.json"),
		}, nil)
}

func TestOutputFormatGitHubCommentWithProjectNamesWithMetadata(t *testing.T) {
	testName := testutil.CalcGoldenFileTestdataDirName()
	GoldenFileCommandTest(t, testName,
		[]string{
			"output",
			"--format", "github-comment",
			"--path", path.Join("./testdata", testName, "infracost.json"),
		}, nil)
}

func TestOutputFormatGitHubCommentMultipleSkipped(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--format", "github-comment", "--path", "./testdata/example_out.json", "--path", "./testdata/terraform_v0.14_breakdown.json", "--path", "./testdata/terraform_v0.14_nochange_breakdown.json", "--path", "./testdata/terraform_v0.14_nochange_breakdown.json"}, nil)
}

func TestOutputFormatGitHubCommentNoChange(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--format", "github-comment", "--path", "./testdata/terraform_v0.14_nochange_breakdown.json"}, nil)
}

func TestOutputFormatGitLabComment(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--format", "gitlab-comment", "--path", "./testdata/example_out.json", "--path", "./testdata/terraform_v0.14_breakdown.json", "--path", "./testdata/terraform_v0.14_nochange_breakdown.json"}, nil)
}

func TestOutputFormatGitLabCommentMultipleSkipped(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--format", "gitlab-comment", "--path", "./testdata/example_out.json", "--path", "./testdata/terraform_v0.14_breakdown.json", "--path", "./testdata/terraform_v0.14_nochange_breakdown.json", "--path", "./testdata/terraform_v0.14_nochange_breakdown.json"}, nil)
}

func TestOutputFormatGitLabCommentNoChange(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--format", "gitlab-comment", "--path", "./testdata/terraform_v0.14_nochange_breakdown.json"}, nil)
}

func TestOutputFormatAzureReposComment(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--format", "azure-repos-comment", "--path", "./testdata/example_out.json", "--path", "./testdata/terraform_v0.14_breakdown.json", "--path", "./testdata/terraform_v0.14_nochange_breakdown.json"}, nil)
}

func TestOutputFormatAzureReposCommentMultipleSkipped(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--format", "azure-repos-comment", "--path", "./testdata/example_out.json", "--path", "./testdata/terraform_v0.14_breakdown.json", "--path", "./testdata/terraform_v0.14_nochange_breakdown.json", "--path", "./testdata/terraform_v0.14_nochange_breakdown.json"}, nil)
}

func TestOutputFormatAzureReposCommentNoChange(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--format", "azure-repos-comment", "--path", "./testdata/terraform_v0.14_nochange_breakdown.json"}, nil)
}

func TestOutputFormatSlackMessage(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--format", "slack-message", "--path", "./testdata/example_out.json", "--path", "./testdata/terraform_v0.14_breakdown.json", "--path", "./testdata/terraform_v0.14_nochange_breakdown.json"}, nil)
}

func TestOutputFormatSlackMessageMultipleSkipped(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--format", "slack-message", "--path", "./testdata/example_out.json", "--path", "./testdata/terraform_v0.14_breakdown.json", "--path", "./testdata/terraform_v0.14_nochange_breakdown.json", "--path", "./testdata/terraform_v0.14_nochange_breakdown.json"}, nil)
}

func TestOutputFormatSlackMessageNoChange(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--format", "slack-message", "--path", "./testdata/terraform_v0.14_nochange_breakdown.json"}, nil)
}

func TestOutputFormatSlackMessageMoreProjects(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--format", "slack-message", "--path", "./testdata/example_out.json", "--path", "./testdata/example_out.json", "--path", "./testdata/example_out.json", "--path", "./testdata/example_out.json", "--path", "./testdata/example_out.json", "--path", "./testdata/example_out.json", "--path", "./testdata/example_out.json"}, nil)
}

func TestOutputFormatTable(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--format", "table", "--path", "./testdata/example_out.json", "--path", "./testdata/azure_firewall_out.json"}, nil)
}

func TestOutputTerraformFieldsAll(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--path", "./testdata/example_out.json", "--path", "./testdata/azure_firewall_out.json", "--fields", "all"}, nil)
}

func TestOutputTerraformOutFileHTML(t *testing.T) {
	testdataName := testutil.CalcGoldenFileTestdataDirName()
	goldenFilePath := "./testdata/" + testdataName + "/infracost_output.golden"
	outputPath := filepath.Join(t.TempDir(), "infracost_output.html")

	GoldenFileCommandTest(t, testdataName, []string{"output", "--path", "./testdata/example_out.json", "--format", "html", "--out-file", outputPath}, nil)

	actual, err := ioutil.ReadFile(outputPath)
	require.Nil(t, err)
	actual = stripDynamicValues(actual)

	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}

func TestOutputTerraformOutFileJSON(t *testing.T) {
	testdataName := testutil.CalcGoldenFileTestdataDirName()
	goldenFilePath := "./testdata/" + testdataName + "/infracost_output.golden"
	outputPath := filepath.Join(t.TempDir(), "infracost_output.json")

	GoldenFileCommandTest(t, testdataName, []string{"output", "--path", "./testdata/example_out.json", "--format", "json", "--out-file", outputPath}, nil)

	actual, err := ioutil.ReadFile(outputPath)
	require.Nil(t, err)
	actual = stripDynamicValues(actual)

	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}

func TestOutputTerraformOutFileTable(t *testing.T) {
	testdataName := testutil.CalcGoldenFileTestdataDirName()
	goldenFilePath := "./testdata/" + testdataName + "/infracost_output.golden"
	outputPath := filepath.Join(t.TempDir(), "infracost_output.txt")

	GoldenFileCommandTest(t, testdataName, []string{"output", "--path", "./testdata/example_out.json", "--out-file", outputPath}, nil)

	actual, err := ioutil.ReadFile(outputPath)
	require.Nil(t, err)
	actual = stripDynamicValues(actual)

	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}

func TestOutputJSONArrayPath(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--path", "[\"./testdata/example_out.json\", \"./testdata/terraform_v0.14*breakdown.json\"]"}, nil)
}
