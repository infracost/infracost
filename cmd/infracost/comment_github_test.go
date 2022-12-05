package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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

func TestCommentGitHubWithNoGuardrailt(t *testing.T) {
	ts := guardrailTestEndpoint(guardrailAddRunResponse{
		GuardrailsChecked: 0,
		Comment:           false,
		Events:            []guardrailEvent{},
	})
	defer ts.Close()

	GuardrailGoldenFileTest(t, testutil.CalcGoldenFileTestdataDirName(), ts.URL)
}

func TestCommentGitHubGuardrailSuccessWithoutComment(t *testing.T) {
	ts := guardrailTestEndpoint(guardrailAddRunResponse{
		GuardrailsChecked: 1,
		Comment:           false,
		Events:            []guardrailEvent{},
	})
	defer ts.Close()

	GuardrailGoldenFileTest(t, testutil.CalcGoldenFileTestdataDirName(), ts.URL)
}

func TestCommentGitHubGuardrailSuccessWithComment(t *testing.T) {
	ts := guardrailTestEndpoint(guardrailAddRunResponse{
		GuardrailsChecked: 1,
		Comment:           true,
		Events:            []guardrailEvent{},
	})
	defer ts.Close()

	GuardrailGoldenFileTest(t, testutil.CalcGoldenFileTestdataDirName(), ts.URL)
}

func TestCommentGitHubGuardrailFailureWithComment(t *testing.T) {
	ts := guardrailTestEndpoint(guardrailAddRunResponse{
		GuardrailsChecked: 1,
		Comment:           true,
		Events: []guardrailEvent{{
			TriggerReason: "Stand by your estimate",
			PrComment:     true,
			BlockPr:       false,
		}},
	})
	defer ts.Close()

	GuardrailGoldenFileTest(t, testutil.CalcGoldenFileTestdataDirName(), ts.URL)
}

func TestCommentGitHubGuardrailFailureWithBlock(t *testing.T) {
	ts := guardrailTestEndpoint(guardrailAddRunResponse{
		GuardrailsChecked: 1,
		Comment:           false,
		Events: []guardrailEvent{{
			TriggerReason: "Stand by your estimate",
			PrComment:     false,
			BlockPr:       true,
		}},
	})
	defer ts.Close()

	GuardrailGoldenFileTest(t, testutil.CalcGoldenFileTestdataDirName(), ts.URL)
}

func TestCommentGitHubGuardrailFailureWithCommentAndBlock(t *testing.T) {
	ts := guardrailTestEndpoint(guardrailAddRunResponse{
		GuardrailsChecked: 1,
		Comment:           true,
		Events: []guardrailEvent{{
			TriggerReason: "Stand by your estimate",
			PrComment:     true,
			BlockPr:       true,
		}},
	})
	defer ts.Close()

	GuardrailGoldenFileTest(t, testutil.CalcGoldenFileTestdataDirName(), ts.URL)
}

func TestCommentGitHubGuardrailFailureWithoutCommentOrBlock(t *testing.T) {
	ts := guardrailTestEndpoint(guardrailAddRunResponse{
		GuardrailsChecked: 1,
		Comment:           false,
		Events: []guardrailEvent{{
			TriggerReason: "Stand by your estimate",
			PrComment:     false,
			BlockPr:       false,
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
	GuardrailsChecked int64            `json:"guardrailsChecked"`
	Comment           bool             `json:"guardrailComment"`
	Events            []guardrailEvent `json:"guardrailEvents"`
}

type guardrailEvent struct {
	TriggerReason string `json:"triggerReason"`
	PrComment     bool   `json:"prComment"`
	BlockPr       bool   `json:"blockPr"`
}

func guardrailTestEndpoint(gr guardrailAddRunResponse) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := ioutil.ReadAll(r.Body)
		graphqlQuery := string(bodyBytes)

		if strings.Contains(graphqlQuery, "mutation($run: RunInput!)") {
			guardrailJson, _ := json.Marshal(gr)

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
