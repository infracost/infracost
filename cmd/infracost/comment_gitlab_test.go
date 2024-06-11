package main_test

import (
	"testing"

	"github.com/infracost/infracost/internal/testutil"
)

func TestCommentGitLabHelp(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"comment", "gitlab", "--help"}, nil)
}

func TestCommentGitLabMergeRequest(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"comment", "gitlab", "--gitlab-token", "abc", "--repo", "test/test", "--merge-request", "5", "--path", "./testdata/terraform_v0.14_breakdown.json", "--dry-run"},
		nil)
}

func TestCommentGitLabCommit(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"comment", "gitlab", "--gitlab-token", "abc", "--repo", "test/test", "--commit", "5", "--path", "./testdata/terraform_v0.14_breakdown.json", "--dry-run"},
		nil)
}

func TestCommentGitLabCommentPath(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"comment", "gitlab", "--gitlab-token", "abc", "--repo", "test/test", "--merge-request", "5", "--path", "./testdata/terraform_v0.14_breakdown.json", "--comment-path", "./testdata/comment.md", "--dry-run"},
		nil)
}
