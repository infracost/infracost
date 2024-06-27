package main_test

import (
	_ "embed"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/infracost/infracost/internal/config"

	"github.com/infracost/infracost/internal/testutil"
)

func TestUpload(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"upload"}, nil)
}

func TestUploadHelp(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"upload", "--help"}, nil)
}

func TestUploadSelfHosted(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer s.Close()

	GoldenFileCommandTest(t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{"upload", "--path", "./testdata/example_out.json"},
		&GoldenFileOptions{CaptureLogs: true},
		func(c *config.RunContext) {
			c.Config.PricingAPIEndpoint = s.URL
		},
	)
}

func TestUploadBadFile(t *testing.T) {
	GoldenFileCommandTest(t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{"upload", "--path", "./testdata/doesnotexist.json", "--log-level", "info"},
		&GoldenFileOptions{CaptureLogs: true},
	)
}

func TestUploadWithPath(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `[{"data": {"addRun":{
			"id":"d92e0196-e5b0-449b-85c9-5733f6643c2f",
			"shareUrl":"",
			"organization":{"id":"767", "name":"tim"}
		}}}]`)
	}))
	defer ts.Close()

	GoldenFileCommandTest(t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{"upload", "--path", "./testdata/example_out.json", "--log-level", "info"},
		&GoldenFileOptions{CaptureLogs: true},
		func(c *config.RunContext) {
			c.Config.DashboardAPIEndpoint = ts.URL
		},
	)
}

func TestUploadWithPathFormatJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `[{"data": {"addRun":{
		  "id": "ff050cb8-eaaa-479f-8865-3fc0fd689b9f",
		  "shareUrl": "",
		  "cloudUrl": "https://dashboard.infracost.io/org/tim/repos/4e936c53-6091-4836-82da-e106ff653aec/runs/ff050cb8-eaaa-479f-8865-3fc0fd689b9f",
		  "pullRequestUrl": "https://dashboard.infracost.io/org/tim/repos/4e936c53-6091-4836-82da-e106ff653aec/pulls/null",
		  "governanceFailures": null,
		  "commentMarkdown": "\n<h4>Governance checks</h4>\n\n<details>\n<summary><strong>🔴 4 failures</strong>...",
		  "governanceResults": [
			{
			  "govType": "tag_policy",
			  "checked": 1,
			  "warnings": [
				"Timtags"
			  ],
			  "failures": [],
			  "unblocked": []
			},
			{
			  "govType": "finops_policy",
			  "checked": 48,
			  "warnings": [
				"Cloudwatch - consider using a retention policy to reduce storage costs",
				"EBS - consider upgrading gp2 volumes to gp3",
				"S3 - consider using a lifecycle policy to reduce storage costs"
			  ],
			  "failures": [],
			  "unblocked": []
			},
			{
			  "govType": "guardrail",
			  "checked": 0,
			  "warnings": [],
			  "failures": [],
			  "unblocked": []
			}
		  ]
		}}}]`)
	}))
	defer ts.Close()

	GoldenFileCommandTest(t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{"upload", "--path", "./testdata/example_out.json", "--log-level", "info", "--format", "json"},
		&GoldenFileOptions{CaptureLogs: true},
		func(c *config.RunContext) {
			c.Config.DashboardAPIEndpoint = ts.URL
		},
	)
}

func TestUploadWithShareLink(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `[{"data": {"addRun":{
			"id":"d92e0196-e5b0-449b-85c9-5733f6643c2f",
			"shareUrl":"http://localhost:3000/share/1234",
			"organization":{"id":"767", "name":"tim"}
		}}}]`)
	}))
	defer ts.Close()

	GoldenFileCommandTest(t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{"upload", "--path", "./testdata/example_out.json", "--log-level", "info"},
		&GoldenFileOptions{CaptureLogs: true},
		func(c *config.RunContext) {
			c.Config.DashboardAPIEndpoint = ts.URL
		},
	)
}

func TestUploadWithCloudDisabled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `[{"data": {"addRun":{
			"id":"d92e0196-e5b0-449b-85c9-5733f6643c2f",
			"shareUrl":"",
			"organization":{"id":"767", "name":"tim"}
		}}}]`)
	}))
	defer ts.Close()

	GoldenFileCommandTest(t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{"upload", "--path", "./testdata/example_out.json", "--log-level", "info"},
		&GoldenFileOptions{CaptureLogs: true},
		func(c *config.RunContext) {
			c.Config.DashboardAPIEndpoint = ts.URL
			f := false
			c.Config.EnableCloud = &f // Should still upload even though we've disabled cloud
		},
	)
}

func TestUploadWithGuardrailSuccess(t *testing.T) {
	ts := governanceTestEndpoint(governanceAddRunResponse{
		CommentMarkdown: "",
		GovernanceResults: []GovernanceResult{{
			Type:    "guardrail",
			Checked: 2,
		}},
	})
	defer ts.Close()

	GoldenFileCommandTest(t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{"upload", "--path", "./testdata/example_out.json", "--log-level", "info"},
		&GoldenFileOptions{CaptureLogs: true},
		func(c *config.RunContext) {
			c.Config.DashboardAPIEndpoint = ts.URL
			f := false
			c.Config.EnableCloud = &f // Should still upload even though we've disabled cloud
		},
	)
}

func TestUploadWithGuardrailFailure(t *testing.T) {
	ts := governanceTestEndpoint(governanceAddRunResponse{
		CommentMarkdown: "",
		GovernanceResults: []GovernanceResult{{
			Type:    "guardrail",
			Checked: 2,
			Warnings: []string{
				"medical problems",
			},
		}},
	})
	defer ts.Close()
	defer ts.Close()

	GoldenFileCommandTest(t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{"upload", "--path", "./testdata/example_out.json", "--log-level", "info"},
		&GoldenFileOptions{CaptureLogs: true},
		func(c *config.RunContext) {
			c.Config.DashboardAPIEndpoint = ts.URL
			f := false
			c.Config.EnableCloud = &f // Should still upload even though we've disabled cloud
		},
	)
}

func TestUploadWithBlockingGuardrailFailure(t *testing.T) {
	ts := governanceTestEndpoint(governanceAddRunResponse{
		CommentMarkdown: "",
		GovernanceResults: []GovernanceResult{{
			Type:    "guardrail",
			Checked: 2,
			Failures: []string{
				"medical problems",
			},
		}},
	})
	defer ts.Close()

	GoldenFileCommandTest(t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{"upload", "--path", "./testdata/example_out.json", "--log-level", "info"},
		&GoldenFileOptions{CaptureLogs: true},
		func(c *config.RunContext) {
			c.Config.DashboardAPIEndpoint = ts.URL
			f := false
			c.Config.EnableCloud = &f // Should still upload even though we've disabled cloud
		},
	)
}

//go:embed testdata/upload_with_blocking_tag_policy_failure/policyResponse.json
var uploadWithBlockingTagPolicyFailureResponse string

func TestUploadWithBlockingTagPolicyFailure(t *testing.T) {
	policyV2Api := GraphqlTestServer(map[string]string{
		"policyResourceAllowList": policyResourceAllowlistGraphQLResponse,
		"evaluatePolicies":        uploadWithBlockingTagPolicyFailureResponse,
	})
	defer policyV2Api.Close()

	dashboardApi := governanceTestEndpoint(governanceAddRunResponse{
		CommentMarkdown: "Tag policy failure",
		GovernanceResults: []GovernanceResult{{
			Type:    "tag_policy",
			Checked: 2,
			Failures: []string{
				"should show as failing",
			},
		}},
	})
	defer dashboardApi.Close()

	GoldenFileCommandTest(t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{"upload", "--path", "./testdata/example_out.json", "--log-level", "info"},
		&GoldenFileOptions{
			CaptureLogs: true,
		},
		func(c *config.RunContext) {
			c.Config.DashboardAPIEndpoint = dashboardApi.URL
			c.Config.PolicyV2APIEndpoint = policyV2Api.URL
			c.Config.PoliciesEnabled = true
		},
	)
}

//go:embed testdata/upload_with_tag_policy_warning/policyResponse.json
var uploadWithTagPolicyWarningResponse string

func TestUploadWithTagPolicyWarning(t *testing.T) {
	policyV2Api := GraphqlTestServer(map[string]string{
		"policyResourceAllowList": policyResourceAllowlistGraphQLResponse,
		"evaluatePolicies":        uploadWithTagPolicyWarningResponse,
	})
	defer policyV2Api.Close()

	dashboardApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `[{"data": {"addRun":{
			"id":"d92e0196-e5b0-449b-85c9-5733f6643c2f",
			"shareUrl":"",
			"organization":{"id":"767", "name":"tim"}
		}}}]`)
	}))
	defer dashboardApi.Close()

	GoldenFileCommandTest(t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{"upload", "--path", "./testdata/example_out.json", "--log-level", "info"},
		&GoldenFileOptions{
			CaptureLogs: true,
		},
		func(c *config.RunContext) {
			c.Config.DashboardAPIEndpoint = dashboardApi.URL
			c.Config.PolicyV2APIEndpoint = policyV2Api.URL
			c.Config.PoliciesEnabled = true
		},
	)
}

//go:embed testdata/upload_with_blocking_fin_ops_policy_failure/policyResponse.json
var uploadWithBlockingFinOpsPolicyFailureResponse string

func TestUploadWithBlockingFinOpsPolicyFailure(t *testing.T) {
	policyV2Api := GraphqlTestServer(map[string]string{
		"policyResourceAllowList": policyResourceAllowlistGraphQLResponse,
		"evaluatePolicies":        uploadWithBlockingFinOpsPolicyFailureResponse,
	})
	defer policyV2Api.Close()

	dashboardApi := governanceTestEndpoint(governanceAddRunResponse{
		CommentMarkdown: "FinOPs policy failure",
		GovernanceResults: []GovernanceResult{{
			Type:    "finops_policy",
			Checked: 2,
			Failures: []string{
				"should show as failing",
			},
		}},
	})
	defer dashboardApi.Close()

	GoldenFileCommandTest(t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{"upload", "--path", "./testdata/example_out.json", "--log-level", "info"},
		&GoldenFileOptions{
			CaptureLogs: true,
		},
		func(c *config.RunContext) {
			c.Config.DashboardAPIEndpoint = dashboardApi.URL
			c.Config.PolicyV2APIEndpoint = policyV2Api.URL
			c.Config.PoliciesEnabled = true
		},
	)
}

//go:embed testdata/upload_with_fin_ops_policy_warning/policyResponse.json
var uploadWithFinOpsPolicyWarningResponse string

func TestUploadWithFinOpsPolicyWarning(t *testing.T) {
	policyV2Api := GraphqlTestServer(map[string]string{
		"policyResourceAllowList": policyResourceAllowlistGraphQLResponse,
		"evaluatePolicies":        uploadWithFinOpsPolicyWarningResponse,
	})
	defer policyV2Api.Close()

	dashboardApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `[{"data": {"addRun":{
			"id":"d92e0196-e5b0-449b-85c9-5733f6643c2f",
			"shareUrl":"",
			"organization":{"id":"767", "name":"tim"}
		}}}]`)
	}))
	defer dashboardApi.Close()

	GoldenFileCommandTest(t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{"upload", "--path", "./testdata/example_out.json", "--log-level", "info"},
		&GoldenFileOptions{
			CaptureLogs: true,
		},
		func(c *config.RunContext) {
			c.Config.DashboardAPIEndpoint = dashboardApi.URL
			c.Config.PolicyV2APIEndpoint = policyV2Api.URL
			c.Config.PoliciesEnabled = true
		},
	)
}
