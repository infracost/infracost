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

func TestUpload(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"upload"}, nil)
}

func TestUploadHelp(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"upload", "--help"}, nil)
}

func TestUploadSelfHosted(t *testing.T) {
	GoldenFileCommandTest(t,
		testutil.CalcGoldenFileTestdataDirName(),
		[]string{"upload", "--path", "./testdata/example_out.json"},
		&GoldenFileOptions{CaptureLogs: true},
		func(c *config.RunContext) {
			c.Config.PricingAPIEndpoint = "https://fake.url"
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
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `[{"data": {"addRun":{
			"id":"d92e0196-e5b0-449b-85c9-5733f6643c2f",
			"shareUrl":"",
			"organization":{"id":"767", "name":"tim"},
			"guardrailsChecked": 2,
            "guardrailComment": false,
            "guardrailEvents": []
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

func TestUploadWithGuardrailFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `[{"data": {"addRun":{
			"id":"d92e0196-e5b0-449b-85c9-5733f6643c2f",
			"shareUrl":"",
			"organization":{"id":"767", "name":"tim"},
			"guardrailsChecked": 2,
            "guardrailComment": false,
            "guardrailEvents": [{
              "triggerReason": "medical problems",
              "prComment": false,
              "blockPr": false,
			}]
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

func TestUploadWithBlockingGuardrailFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `[{"data": {"addRun":{
			"id":"d92e0196-e5b0-449b-85c9-5733f6643c2f",
			"shareUrl":"",
			"organization":{"id":"767", "name":"tim"},
			"guardrailsChecked": 2,
            "guardrailComment": false,
            "guardrailEvents": [{
              "triggerReason": "medical problems",
              "prComment": false,
              "blockPr": true,
			}]
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

//go:embed testdata/upload_with_blocking_tag_policy_failure/tagPolicyResponse.json
var uploadWithBlockingTagPolicyFailureResponse string

func TestUploadWithBlockingTagPolicyFailure(t *testing.T) {
	tagPolicyApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, uploadWithBlockingTagPolicyFailureResponse)
	}))
	defer tagPolicyApi.Close()

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
			c.Config.TagPolicyAPIEndpoint = tagPolicyApi.URL
		},
	)
}

//go:embed testdata/upload_with_tag_policy_warning/tagPolicyResponse.json
var uploadWithTagPolicyWarningResponse string

func TestUploadWithTagPolicyWarning(t *testing.T) {
	tagPolicyApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, uploadWithTagPolicyWarningResponse)
	}))
	defer tagPolicyApi.Close()

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
			c.Config.TagPolicyAPIEndpoint = tagPolicyApi.URL
		},
	)
}
