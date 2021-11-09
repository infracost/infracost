package main_test

import (
	"github.com/infracost/infracost/internal/testutil"
	"testing"
)

func TestOutputHelp(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--help"}, nil)
}

func TestOutputFormatHTML(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--format", "html", "--path", "./testdata/example_out.json", "--path", "./testdata/azure_firewall_out.json"}, nil)
}

func TestOutputFormatJSON(t *testing.T) {
	opts := DefaultOptions()
	opts.IsJSON = true
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--format", "json", "--path", "./testdata/example_out.json", "--path", "./testdata/azure_firewall_out.json"}, opts)
}

func TestOutputFormatGithubComment(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--format", "github-comment", "--path", "./testdata/example_out.json", "--path", "./testdata/azure_firewall_out.json"}, nil)
}

func TestOutputFormatTable(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--format", "table", "--path", "./testdata/example_out.json", "--path", "./testdata/azure_firewall_out.json"}, nil)
}

func TestOutputTerraformFieldsAll(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"output", "--path", "./testdata/example_out.json", "--path", "./testdata/azure_firewall_out.json", "--fields", "all"}, nil)
}
