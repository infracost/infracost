package main_test

import (
	"testing"

	"github.com/infracost/infracost/internal/testutil"
)

func TestCommentGitHubHelp(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"comment", "github", "--help"}, nil)
}

func TestCommentGitHubPullRequest(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"comment", "github", "--github-token", "abc", "--repo", "test/test", "--pull-request", "5", "--path", "./testdata/terraform_v0.14_breakdown.json", "--dry-run"},
		nil)
}

func TestCommentGitHubCommit(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"comment", "github", "--github-token", "abc", "--repo", "test/test", "--commit", "5", "--path", "./testdata/terraform_v0.14_breakdown.json", "--dry-run"},
		nil)
}
