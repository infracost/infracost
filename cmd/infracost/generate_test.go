package main_test

import (
	"path"
	"testing"

	"github.com/infracost/infracost/internal/testutil"
)

func TestGenerateConfig(t *testing.T) {
	dir := testutil.CalcGoldenFileTestdataDirName()
	GoldenFileCommandTest(
		t,
		dir,
		[]string{
			"generate",
			"config",
			"--template-path",
			path.Join("./testdata", dir, "infracost.yml.tmpl"),
			"--repo-path",
			path.Join("./testdata", dir),
		},
		nil)
}

func TestGenerateConfigWarning(t *testing.T) {
	dir := testutil.CalcGoldenFileTestdataDirName()
	GoldenFileCommandTest(
		t,
		dir,
		[]string{
			"generate",
			"config",
			"--template-path",
			path.Join("./testdata", dir, "infracost.yml.tmpl"),
			"--repo-path",
			path.Join("./testdata", dir),
		},
		nil)
}

func TestGenerateConfigDetectedProjects(t *testing.T) {
	dir := testutil.CalcGoldenFileTestdataDirName()
	GoldenFileCommandTest(
		t,
		dir,
		[]string{
			"generate",
			"config",
			"--template-path",
			path.Join("./testdata", dir, "infracost.yml.tmpl"),
			"--repo-path",
			path.Join("./testdata", dir),
		},
		nil)
}

func TestGenerateConfigAutoDetect(t *testing.T) {
	dir := testutil.CalcGoldenFileTestdataDirName()
	GoldenFileCommandTest(
		t,
		dir,
		[]string{
			"generate",
			"config",
			"--repo-path",
			path.Join("./testdata", dir),
		},
		nil)
}

func TestGenerateConfigAutoDetectWithNestedTfvars(t *testing.T) {
	dir := testutil.CalcGoldenFileTestdataDirName()
	GoldenFileCommandTest(
		t,
		dir,
		[]string{
			"generate",
			"config",
			"--repo-path",
			path.Join("./testdata", dir),
		},
		nil)
}
