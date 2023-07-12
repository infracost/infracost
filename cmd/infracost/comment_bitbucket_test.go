package main_test

import (
	_ "embed"
	"fmt"
	"github.com/infracost/infracost/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/infracost/infracost/internal/testutil"
)

func TestCommentBitbucketHelp(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"comment", "bitbucket", "--help"}, nil)
}

func TestCommentBitbucketPullRequest(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"comment", "bitbucket", "--bitbucket-token", "abc", "--repo", "test/test", "--pull-request", "5", "--path", "./testdata/terraform_v0.14_breakdown.json", "--dry-run"},
		nil)
}

func TestCommentBitbucketCommit(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"comment", "bitbucket", "--bitbucket-token", "abc", "--repo", "test/test", "--commit", "5", "--path", "./testdata/terraform_v0.14_breakdown.json", "--dry-run"},
		nil)
}

func TestCommentBitbucketExcludeDetails(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"comment", "bitbucket", "--bitbucket-token", "abc", "--repo", "test/test", "--pull-request", "5", "--path", "./testdata/terraform_v0.14_breakdown.json", "--exclude-cli-output", "--dry-run"},
		nil)
}

//go:embed testdata/comment_bitbucket_with_tag_policy_checks/tagPolicyResponse.json
var commentBitbucketWithTagPolicyChecksTagPolicyResponse string

func TestCommentBitbucketWithTagPolicyChecks(t *testing.T) {
	tagPolicyApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, commentBitbucketWithTagPolicyChecksTagPolicyResponse)
	}))
	defer tagPolicyApi.Close()

	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"comment",
			"bitbucket",
			"--bitbucket-token", "abc",
			"--repo", "test/test",
			"--pull-request", "5",
			"--path", "./testdata/changes.json",
			"--dry-run"},
		&GoldenFileOptions{
			Env: map[string]string{
				"INFRACOST_TAG_POLICY_API_ENDPOINT": tagPolicyApi.URL,
			},
		},
		func(c *config.RunContext) {
			t := true
			c.Config.EnableCloudUpload = &t
		},
	)
}
