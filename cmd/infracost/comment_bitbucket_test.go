package main_test

import (
	_ "embed"
	"testing"

	"github.com/infracost/infracost/internal/config"

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

//go:embed testdata/comment_bitbucket_with_tag_policy_checks/policyV2Response.json
var commentBitbucketWithTagPolicyChecksPolicyV2Response string

func TestCommentBitbucketWithTagPolicyChecks(t *testing.T) {
	policyV2Api := GraphqlTestServer(map[string]string{
		"policyResourceAllowList": policyResourceAllowlistGraphQLResponse,
		"evaluatePolicies":        commentBitbucketWithTagPolicyChecksPolicyV2Response,
	})
	defer policyV2Api.Close()

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
				"INFRACOST_POLICY_V2_API_ENDPOINT": policyV2Api.URL,
			},
		},
		func(c *config.RunContext) {
			t := true
			c.Config.EnableCloudUpload = &t
		},
	)
}

//go:embed testdata/comment_bitbucket_with_fin_ops_policy_checks/policyV2Response.json
var commentBitbucketWithFinOpsPolicyChecksPolicyV2Response string

func TestCommentBitbucketWithFinOpsPolicyChecks(t *testing.T) {
	policyV2Api := GraphqlTestServer(map[string]string{
		"policyResourceAllowList": policyResourceAllowlistGraphQLResponse,
		"evaluatePolicies":        commentBitbucketWithFinOpsPolicyChecksPolicyV2Response,
	})
	defer policyV2Api.Close()

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
				"INFRACOST_POLICY_V2_API_ENDPOINT": policyV2Api.URL,
			},
		},
		func(c *config.RunContext) {
			t := true
			c.Config.EnableCloudUpload = &t
		},
	)
}
