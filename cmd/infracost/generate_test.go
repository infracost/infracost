package main_test

import (
	"path"
	"path/filepath"
	"strings"
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

// To write `generate` golden file tests, use a tool like
// https://tree.nathanfriend.io/ to write a `tree.txt` file which gives an
// example of the test case you want to cover. This should be placed in a
// subdirectory in the `./testdata/generate` directory. Each directory under the
// `./testdata/generate` directory with a `tree.txt` file will be run by
// TestGenerateConfigAutoDetect as a subtest.
//
// Each test will read the `tree.txt` and create a shallow representation of it
// in a tmp directory. It will then run the `infracost generate config` command
// on this directory and compare the output with any output saved in
// `expected.golden.
//
// All files written to the temp directory before running `generate` will have
// blank contents unless they contain the following naming conventions:
//
//   - main.tf => will generate a file with a provider block
//   - backend.tf => will generate a file with a terraform block with a backend
//     child block
func TestGenerateConfigAutoDetect(t *testing.T) {
	dirs := testutil.FindDirectoriesWithTreeFile(t, "./testdata/generate")
	for _, dir := range dirs {
		t.Run(strings.Join(strings.Split(filepath.Base(dir), "_"), " "), func(tt *testing.T) {
			tempDir := tt.TempDir()
			testutil.CreateDirectoryStructure(tt, path.Join(dir, "tree.txt"), tempDir)

			actual := GetCommandOutput(tt, []string{
				"generate",
				"config",
				"--repo-path",
				tempDir,
			}, nil)

			testutil.AssertGoldenFile(tt, path.Join(dir, "expected.golden"), actual)
		})
	}
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
