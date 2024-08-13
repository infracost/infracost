package main_test

import (
	"regexp"
	"testing"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/testutil"
)

func TestFlagErrorsNoPath(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown"}, nil)
}

func TestFlagErrorsPathAndConfigFile(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "./testdata/example_plan.json", "--usage-file", "./testdata/example_usage.yml", "--config-file", "infracost-config.yml"}, nil)
}

func TestFlagErrorsConfigFileAndTerraformWorkspace(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--config-file", "./testdata/infracost-config.yml", "--terraform-workspace", "dev"}, nil)
}

func TestFlagErrorsConfigFileAndTerraformWorkspaceEnv(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--config-file", "./testdata/infracost-config.yml",
		},
		&GoldenFileOptions{
			Env: map[string]string{
				"INFRACOST_TERRAFORM_WORKSPACE": "dev",
			},
		})
}

func TestConfigFileNilProjectsErrors(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--config-file", "./testdata/infracost-config-nil-projects.yml"}, nil)
}

func TestConfigFileInvalidKeysErrors(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--config-file", "./testdata/infracost-config-invalid-key.yml"}, nil)
}

func TestConfigFileInvalidPathErrors(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--config-file", "./testdata/infracost-config-invalid-path.yml"}, nil)
}

func TestFlagErrorsTerraformWorkspaceFlagAndEnv(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(),
		[]string{
			"breakdown",
			"--path", "../../examples/terraform",
			"--terraform-workspace", "prod",
		},
		&GoldenFileOptions{
			Env: map[string]string{
				"INFRACOST_TERRAFORM_WORKSPACE": "dev",
			},
		})
}

func TestCatchesRuntimeError(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "../../examples/terraform", "--terraform-workspace", "prod"}, &GoldenFileOptions{CaptureLogs: true}, func(c *config.RunContext) {
		// this should blow up the application
		c.Config.Projects = []*config.Project{nil, nil}
	})
}

func TestConfigFileWithYorConfig(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--config-file", "./testdata/infracost-config-yor.yml", "--format", "json"}, &GoldenFileOptions{
		IsJSON:      true,
		JSONInclude: regexp.MustCompile("^(name|tags)$"),
		JSONExclude: regexp.MustCompile("^(path)$"),
	})
}

func TestConfigFileWithYorEnv(t *testing.T) {
	t.Setenv("YOR_SIMPLE_TAGS", `{"B": "another-override"}`)
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--config-file", "./testdata/infracost-config-yor.yml", "--format", "json"}, &GoldenFileOptions{
		IsJSON:      true,
		JSONInclude: regexp.MustCompile("^(name|tags)$"),
		JSONExclude: regexp.MustCompile("^(path)$"),
	})
}
