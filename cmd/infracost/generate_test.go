package main_test

import (
	"bytes"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

		t.Run(name, func(tt *testing.T) {
			tempDir := tt.TempDir()
			testutil.CreateDirectoryStructure(tt, path.Join(dir, "tree.txt"), tempDir)

			args := []string{
				"generate",
				"config",
				"--repo-path",
				tempDir,
			}

			templatePath := path.Join(dir, "infracost.yml.tmpl")
			if _, err := os.Stat(templatePath); err == nil {
				tt.Run("template-path", func(tm *testing.T) {
					assertGoldenFile(tm, append(args, "--template-path", templatePath), dir)
				})

				tt.Run("template", func(tm *testing.T) {
					b, err := os.ReadFile(templatePath)
					assert.NoError(tm, err)
					assertGoldenFile(tm, append(args, "--template", string(b)), dir)
				})

				return
			} else {
				assertGoldenFile(tt, args, dir)
			}
		})

		t.Run(name+"-tree", func(tt *testing.T) {
			tempDir := tt.TempDir()
			testutil.CreateDirectoryStructure(tt, path.Join(dir, "tree.txt"), tempDir)

			args := []string{
				"generate",
				"config",
				"--repo-path",
				tempDir,
				"--tree-file",
				path.Join(dir, "infracost-tree.txt"),
			}
			templatePath := path.Join(dir, "infracost.yml.tmpl")
			if _, err := os.Stat(templatePath); err == nil {
				args = append(args, "--template-path", templatePath)
			}

			assertGoldenFileTree(tt, args, dir)
		})
	}
}

func assertGoldenFileTree(tt *testing.T, args []string, dir string) {
	GetCommandOutput(tt, args, nil)

	actual, err := os.ReadFile(path.Join(dir, "infracost-tree.txt"))
	require.NoError(tt, err)

	if _, err := os.Stat(path.Join(dir, "expected-tree.golden")); err != nil {
		err := os.WriteFile(path.Join(dir, "expected-tree.golden"), actual, 0600)
		assert.NoError(tt, err)
		return
	}

	testutil.AssertGoldenFile(tt, path.Join(dir, "expected-tree.golden"), actual)
}

func assertGoldenFile(tt *testing.T, args []string, dir string) {
	buf := bytes.NewBuffer([]byte{})
	actual := GetCommandOutput(tt, args, nil, func(ctx *config.RunContext) {
		ctx.Config.SetLogWriter(buf)
		ctx.Config.LogLevel = "debug"

		err := logging.ConfigureBaseLogger(ctx.Config)
		assert.NoError(tt, err)
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

		err := os.WriteFile(path.Join(dir, "actual.txt"), out.Bytes(), os.ModePerm) // nolint: gosec
		assert.NoError(tt, err)
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
