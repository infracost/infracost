package main_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
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

func TestCommentGitHubShowAllProjects(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"comment", "github", "--github-token", "abc", "--repo", "test/test", "--commit", "5", "--show-all-projects", "--path", "./testdata/terraform_v0.14_breakdown.json", "--path", "./testdata/terraform_v0.14_nochange_breakdown.json", "--dry-run"},
		nil)
}

func TestCommentGitHubShowChangedProjects(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"comment", "github", "--github-token", "abc", "--repo", "test/test", "--commit", "5", "--show-changed", "--path", "./testdata/changes.json", "--dry-run"},
		nil)
}

func TestCommentGitHubWithMissingGuardrailCheckPath(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"comment",
			"github",
			"--github-token", "abc",
			"--repo", "test/test",
			"--commit", "5",
			"--show-changed",
			"--path", "./testdata/changes.json",
			"--dry-run",
			"--additional-path", "./does/not/exist.md"},
		nil)
}

func TestCommentGitHubWithGuardrailCheckPath(t *testing.T) {
	dir := path.Join("./testdata", testutil.CalcGoldenFileTestdataDirName())
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"comment",
			"github",
			"--github-token", "abc",
			"--repo", "test/test",
			"--commit", "5",
			"--show-changed",
			"--path", "./testdata/changes.json",
			"--dry-run",
			"--additional-path", path.Join(dir, "./additional-comment.md")},
		nil)
}

//go:embed testdata/comment_git_hub_with_fin_ops_policy_checks/policyV2Response.json
var commentGitHubWithFinOpsPolicyChecksTagPolicyResponse string

func TestCommentGitHubWithFinOpsPolicyChecks(t *testing.T) {
	policyV2Api := GraphqlTestServer(map[string]string{
		"policyResourceAllowList": policyResourceAllowlistGraphQLResponse,
		"evaluatePolicies":        commentGitHubWithFinOpsPolicyChecksTagPolicyResponse,
	})
	defer policyV2Api.Close()

	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"comment",
			"github",
			"--github-token", "abc",
			"--repo", "test/test",
			"--commit", "5",
			"--show-changed",
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

//go:embed testdata/comment_git_hub_with_tag_policy_checks/policyV2Response.json
var commentGitHubWithTagPolicyChecksTagPolicyResponse string

func TestCommentGitHubWithTagPolicyChecks(t *testing.T) {
	policyV2Api := GraphqlTestServer(map[string]string{
		"policyResourceAllowList": policyResourceAllowlistGraphQLResponse,
		"evaluatePolicies":        commentGitHubWithTagPolicyChecksTagPolicyResponse,
	})
	defer policyV2Api.Close()

	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"comment",
			"github",
			"--github-token", "abc",
			"--repo", "test/test",
			"--commit", "5",
			"--show-changed",
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

func TestCommentGitHubWithTagPolicyCheckMerge(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"comment",
			"github",
			"--github-token", "abc",
			"--repo", "test/test",
			"--commit", "5",
			"--show-changed",
			"--path", "./testdata/comment_git_hub_with_tag_policy_check_merge/infracost1.json",
			"--path", "./testdata/comment_git_hub_with_tag_policy_check_merge/infracost2.json",
			"--dry-run"},
		nil,
		func(c *config.RunContext) {
			t := true
			c.Config.EnableCloudUpload = &t
		},
	)
}

var ghZeroCommentsResponse = `{ "data": { "repository": { "pullRequest": { "comments": { "nodes": [], "pageInfo": { "endCursor": "abc", "hasNextPage": false }}}}}}`
var ghOneMatchingCommentResponse = `{ "data": { "repository": { "pullRequest": { "comments": { "nodes": [ 
            { "id": "123", "body": "infracomment body here, followed by tag: [//]: <> (infracost-comment)" }
          ], "pageInfo": { "endCursor": "abc", "hasNextPage": false }}}}}}`

func TestCommentGitHubSkipNoDiffWithoutInitialComment(t *testing.T) {
	githubGraphQLresponses := []string{
		// show zero comments in the response for findComments
		ghZeroCommentsResponse,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, githubGraphQLresponses[0])
		githubGraphQLresponses = githubGraphQLresponses[1:]
	}))
	defer ts.Close()

	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"comment",
			"github", "--github-token",
			"abc", "--repo", "test/test",
			"--pull-request", "5",
			"--path", "./testdata/terraform_v0.14_nochange_breakdown.json",
			"--skip-no-diff",
			"--log-level", "info",
			"--github-api-url", ts.URL},
		&GoldenFileOptions{CaptureLogs: true},
	)
}

func TestCommentGitHubSkipNoDiffWithInitialComment(t *testing.T) {
	githubGraphQLresponses := []string{
		// show one comment with a matching tag in the response for findComments
		ghOneMatchingCommentResponse,
		// empty json as response to the post comment mutation
		`{}`,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, githubGraphQLresponses[0])
		githubGraphQLresponses = githubGraphQLresponses[1:]
	}))
	defer ts.Close()

	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"comment",
			"github", "--github-token",
			"abc", "--repo", "test/test",
			"--pull-request", "5",
			"--path", "./testdata/terraform_v0.14_nochange_breakdown.json",
			"--skip-no-diff",
			"--log-level", "info",
			"--github-api-url", ts.URL},
		&GoldenFileOptions{CaptureLogs: true},
	)
}

func TestCommentGitHubNewAndHideSkipNoDiffWithoutInitialComment(t *testing.T) {
	githubGraphQLresponses := []string{
		// show zero comments in the response for findComments
		ghZeroCommentsResponse,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, githubGraphQLresponses[0])
		githubGraphQLresponses = githubGraphQLresponses[1:]
	}))
	defer ts.Close()

	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"comment",
			"github", "--github-token",
			"abc", "--repo", "test/test",
			"--pull-request", "5",
			"--path", "./testdata/terraform_v0.14_nochange_breakdown.json",
			"--behavior", "hide-and-new",
			"--skip-no-diff",
			"--log-level", "info",
			"--github-api-url", ts.URL},
		&GoldenFileOptions{CaptureLogs: true},
	)
}

func TestCommentGitHubNewAndHideSkipNoDiffWithInitialComment(t *testing.T) {
	githubGraphQLresponses := []string{
		// show one comment with a matching tag in the response for findComments
		ghOneMatchingCommentResponse,
		`{}`, // empty json as response to the hide comment mutation
		`{}`, // empty json as response to the post comment mutation
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, githubGraphQLresponses[0])
		githubGraphQLresponses = githubGraphQLresponses[1:]
	}))
	defer ts.Close()

	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"comment",
			"github", "--github-token",
			"abc", "--repo", "test/test",
			"--pull-request", "5",
			"--path", "./testdata/terraform_v0.14_nochange_breakdown.json",
			"--behavior", "hide-and-new",
			"--skip-no-diff",
			"--log-level", "info",
			"--github-api-url", ts.URL},
		&GoldenFileOptions{CaptureLogs: true},
	)
}

func TestCommentGitHubDeleteAndNewSkipNoDiffWithoutInitialComment(t *testing.T) {
	githubGraphQLresponses := []string{
		// show zero comments in the response for findComments
		ghZeroCommentsResponse,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, githubGraphQLresponses[0])
		githubGraphQLresponses = githubGraphQLresponses[1:]
	}))
	defer ts.Close()

	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"comment",
			"github", "--github-token",
			"abc", "--repo", "test/test",
			"--pull-request", "5",
			"--path", "./testdata/terraform_v0.14_nochange_breakdown.json",
			"--behavior", "delete-and-new",
			"--skip-no-diff",
			"--log-level", "info",
			"--github-api-url", ts.URL},
		&GoldenFileOptions{CaptureLogs: true},
	)
}

func TestCommentGitHubDeleteAndNewSkipNoDiffWithInitialComment(t *testing.T) {
	githubGraphQLresponses := []string{
		// show one comment with a matching tag in the response for findComments
		ghOneMatchingCommentResponse,
		`{}`, // empty json as response to the delete comment mutation
		`{}`, // empty json as response to the post comment mutation
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, githubGraphQLresponses[0])
		githubGraphQLresponses = githubGraphQLresponses[1:]
	}))
	defer ts.Close()

	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{"comment",
			"github", "--github-token",
			"abc", "--repo", "test/test",
			"--pull-request", "5",
			"--path", "./testdata/terraform_v0.14_nochange_breakdown.json",
			"--behavior", "delete-and-new",
			"--skip-no-diff",
			"--log-level", "info",
			"--github-api-url", ts.URL},
		&GoldenFileOptions{CaptureLogs: true},
	)
}

func TestCommentGitHubWithNoGuardrailt(t *testing.T) {
	ts := guardrailTestEndpoint(guardrailAddRunResponse{
		GovernanceComment: "",
		GovernanceResults: []GovernanceResult{{
			Type:    "guardrail",
			Checked: 0,
		}},
	})
	defer ts.Close()

	GuardrailGoldenFileTest(t, testutil.CalcGoldenFileTestdataDirName(), ts.URL)
}

func TestCommentGitHubGuardrailSuccessWithoutComment(t *testing.T) {
	ts := guardrailTestEndpoint(guardrailAddRunResponse{
		GovernanceComment: "",
		GovernanceResults: []GovernanceResult{{
			Type:    "guardrail",
			Checked: 1,
		}},
	})
	defer ts.Close()

	GuardrailGoldenFileTest(t, testutil.CalcGoldenFileTestdataDirName(), ts.URL)
}

func TestCommentGitHubGuardrailSuccessWithComment(t *testing.T) {
	ts := guardrailTestEndpoint(guardrailAddRunResponse{
		GovernanceComment: "<p><strong>✅ Guardrails passed</strong></p>",
		GovernanceResults: []GovernanceResult{{
			Type:    "guardrail",
			Checked: 1,
		}},
	})
	defer ts.Close()

	GuardrailGoldenFileTest(t, testutil.CalcGoldenFileTestdataDirName(), ts.URL)
}

func TestCommentGitHubGuardrailFailureWithComment(t *testing.T) {
	ts := guardrailTestEndpoint(guardrailAddRunResponse{
		GovernanceComment: `<details>
<summary><strong>⚠️ Guardrails triggered</strong></summary>
				
> - <b>Warning</b>: Stand by your estimate
</details>`,
		GovernanceResults: []GovernanceResult{{
			Type:    "guardrail",
			Checked: 1,
			Warnings: []string{
				"Stand by your estimate",
			},
		}},
	})
	defer ts.Close()

	GuardrailGoldenFileTest(t, testutil.CalcGoldenFileTestdataDirName(), ts.URL)
}

func TestCommentGitHubGuardrailFailureWithBlock(t *testing.T) {
	ts := guardrailTestEndpoint(guardrailAddRunResponse{
		GovernanceComment: "",
		GovernanceResults: []GovernanceResult{{
			Type:      "guardrail",
			Failures:  []string{"Stand by your estimate"},
			Unblocked: []string{"Unblocked ge"},
			Checked:   1,
			// Comment:           false,
		}},
	})
	defer ts.Close()

	GuardrailGoldenFileTest(t, testutil.CalcGoldenFileTestdataDirName(), ts.URL)
}

func TestCommentGitHubGuardrailFailureWithCommentAndBlock(t *testing.T) {
	ts := guardrailTestEndpoint(guardrailAddRunResponse{
		GovernanceComment: `<details>
<summary><strong>❌ Guardrails triggered (needs action)</strong></summary>
This change is blocked, either reduce the costs or wait for an admin to review and unblock it.
				
> - <b>Blocked</b>: Stand by your estimate
</details>`,
		GovernanceResults: []GovernanceResult{{
			Type:    "guardrail",
			Checked: 1,
			Failures: []string{
				"Stand by your estimate",
			},
		}},
	})
	defer ts.Close()

	GuardrailGoldenFileTest(t, testutil.CalcGoldenFileTestdataDirName(), ts.URL)
}

func TestCommentGitHubGuardrailFailureWithoutCommentOrBlock(t *testing.T) {
	ts := guardrailTestEndpoint(guardrailAddRunResponse{
		GovernanceComment: "",
		GovernanceResults: []GovernanceResult{{
			Type:    "guardrail",
			Checked: 1,
			Warnings: []string{
				"Stand by your estimate",
			},
		}},
	})
	defer ts.Close()

	GuardrailGoldenFileTest(t, testutil.CalcGoldenFileTestdataDirName(), ts.URL)
}

// helpers

func GuardrailGoldenFileTest(t *testing.T, testName, guardrailEndpointUrl string) {
	GoldenFileCommandTest(
		t,
		testName,
		[]string{
			"comment", "github",
			"--behavior", "new",
			"--github-token", "abc",
			"--repo", "test/test",
			"--pull-request", "5",
			"--path", "./testdata/terraform_v0.14_breakdown.json",
			"--log-level", "info",
			"--github-api-url", guardrailEndpointUrl,
		},
		&GoldenFileOptions{CaptureLogs: true},
		func(c *config.RunContext) {
			c.Config.DashboardAPIEndpoint = guardrailEndpointUrl
			t := true
			c.Config.EnableCloud = &t
		},
	)
}

type guardrailAddRunResponse struct {
	GovernanceComment string             `json:"governanceComment"`
	GovernanceResults []GovernanceResult `json:"governanceResults"`
}

type GovernanceResult struct {
	Type      string   `json:"govType"`
	Checked   int64    `json:"checked"`
	Warnings  []string `json:"warnings"`
	Failures  []string `json:"failures"`
	Unblocked []string `json:"unblocked"`
}

func guardrailTestEndpoint(garr guardrailAddRunResponse) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		graphqlQuery := string(bodyBytes)

		if strings.Contains(graphqlQuery, "mutation($run: RunInput!)") {
			guardrailJson, _ := json.Marshal(garr)

			fmt.Fprintf(w, `[{"data": {"addRun":
				%v
			}}]\n`, string(guardrailJson))
		} else {
			for _, s := range strings.Split(string(bodyBytes), "details") {
				if strings.Contains(s, "Guardrail") {
					logging.Logger.Warn(s)
				}
			}
			fmt.Fprintln(w, "")
		}
	}))
}
