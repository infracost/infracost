package main_test

import (
	"bytes"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
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
		name := filepath.Base(dir)
		//if name != "sibling_and_child_default" {
		//	continue
		//}

		t.Run(name, func(tt *testing.T) {

			tempDir := tt.TempDir()
			testutil.CreateDirectoryStructure(tt, path.Join(dir, "tree.txt"), tempDir)

			buf := bytes.NewBuffer([]byte{})
			actual := GetCommandOutput(tt, []string{
				"generate",
				"config",
				"--repo-path",
				tempDir,
			}, nil, func(ctx *config.RunContext) {
				ctx.Config.SetLogWriter(buf)
				ctx.Config.LogLevel = "debug"

				logging.ConfigureBaseLogger(ctx.Config)
			})

			equal := testutil.AssertGoldenFile(tt, path.Join(dir, "expected.golden"), actual)
			// if the expected is not equal to the actual output let's write the actual to a
			// file in the folder (this is git ignored) so that we can do easier comparisons
			// about the missing/incorrect data.
			if !equal {
				out := bytes.NewBuffer(actual)
				out.Write([]byte("\n\n"))
				i := buf.Bytes()
				out.Write(i)

				err := os.WriteFile(path.Join(dir, "actual.txt"), out.Bytes(), os.ModePerm)
				assert.NoError(tt, err)
			}
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
