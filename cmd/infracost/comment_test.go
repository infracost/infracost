package main_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/infracost/infracost/internal/testutil"
)

func TestComment(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"comment"}, nil)
}

func TestCommentHelp(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"comment", "--help"}, nil)
}

func TestCommentBackoffRetry(t *testing.T) {
	var attempts int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		attempts += 1
		assert.Equal(t, "/api/v3/repos/infracost/infracost/issues/8/comments", r.RequestURI)
		if (attempts % 3) < 2 {
			w.WriteHeader(400)
			return
		}

		fmt.Fprintf(w, `{
  "id": 1,
  "node_id": "MDEyOklzc3VlQ29tbWVudDE=",
  "url": "https://api.github.com/repos/infracost/infracost/issues/comments/1",
  "html_url": "https://github.com/infracost/infracost/issues/1347#issuecomment-1",
  "body": "Me too",
  "user": {
    "login": "infracost",
    "id": 1,
    "node_id": "MDQ6VXNlcjE=",
    "avatar_url": "https://github.com/images/error/octocat_happy.gif",
    "gravatar_id": "",
    "url": "https://api.github.com/users/infracost",
    "html_url": "https://github.com/infracost",
    "followers_url": "https://api.github.com/users/infracost/followers",
    "following_url": "https://api.github.com/users/infracost/following{/other_user}",
    "gists_url": "https://api.github.com/users/infracost/gists{/gist_id}",
    "starred_url": "https://api.github.com/users/infracost/starred{/owner}{/repo}",
    "subscriptions_url": "https://api.github.com/users/infracost/subscriptions",
    "organizations_url": "https://api.github.com/users/infracost/orgs",
    "repos_url": "https://api.github.com/users/infracost/repos",
    "events_url": "https://api.github.com/users/infracost/events{/privacy}",
    "received_events_url": "https://api.github.com/users/infracost/received_events",
    "type": "User",
    "site_admin": false
  },
  "created_at": "2011-04-14T16:00:49Z",
  "updated_at": "2011-04-14T16:00:49Z",
  "issue_url": "https://api.github.com/repos/infracost/infracost/issues/1347",
  "author_association": "COLLABORATOR"
}`)
	}))
	defer ts.Close()

	dir := testutil.CalcGoldenFileTestdataDirName()
	GoldenFileCommandTest(t, dir, []string{
		"comment",
		"github",
		"--github-api-url", ts.URL,
		"--github-token", "test-token",
		"--pull-request", "8",
		"--behavior", "new",
		"--path", path.Join("./testdata", dir, "infracost.json"),
		"--repo", "infracost/infracost",
	}, nil)

	assert.Equal(t, 2, attempts%3)
}
