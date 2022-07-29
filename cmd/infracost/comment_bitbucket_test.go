package main_test

import (
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
