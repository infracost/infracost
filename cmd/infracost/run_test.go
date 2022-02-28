package main_test

import (
	"os"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	main "github.com/infracost/infracost/cmd/infracost"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
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
	os.Setenv("INFRACOST_TERRAFORM_WORKSPACE", "dev")
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--config-file", "./testdata/infracost-config.yml"}, nil)
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
	os.Setenv("INFRACOST_TERRAFORM_WORKSPACE", "dev")
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "../../examples/terraform", "--terraform-workspace", "prod"}, nil)
}

func TestCatchesRuntimeError(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"breakdown", "--path", "../../examples/terraform", "--terraform-workspace", "prod"}, &GoldenFileOptions{CaptureLogs: true}, func(c *config.RunContext) {
		// this should blow up the application
		c.Config.Projects = []*config.Project{nil, nil}
	})
}

func TestAddHCLEnvVars(t *testing.T) {
	type args struct {
		r      output.Root
		hclR   output.Root
		osVars map[string]string
		env    map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "test nil hcl total monthly",
			args: args{
				r: output.Root{
					TotalMonthlyCost: decimalPtr(decimal.NewFromFloat(1.993)),
				},
				hclR: output.Root{},
				env:  map[string]interface{}{},
			},
			want: map[string]interface{}{
				"hclPercentChange":    "-100.00",
				"absHclPercentChange": "100.00",
			},
		},
		{
			name: "test nil total monthly",
			args: args{
				r:    output.Root{},
				hclR: output.Root{},
				env:  map[string]interface{}{},
			},
			want: map[string]interface{}{
				"hclPercentChange":    "0.00",
				"absHclPercentChange": "0.00",
			},
		},
		{
			name: "test correctly computes percent",
			args: args{
				r: output.Root{
					TotalMonthlyCost: decimalPtr(decimal.NewFromInt(10)),
				},
				hclR: output.Root{
					TotalMonthlyCost: decimalPtr(decimal.NewFromInt(8)),
				},
				env: map[string]interface{}{},
			},
			want: map[string]interface{}{
				"hclPercentChange":    "-20.00",
				"absHclPercentChange": "20.00",
			},
		},
		{
			name: "test correctly formats percent",
			args: args{
				r: output.Root{
					TotalMonthlyCost: decimalPtr(decimal.NewFromInt(11)),
				},
				hclR: output.Root{
					TotalMonthlyCost: decimalPtr(decimal.NewFromInt(7)),
				},
				env: map[string]interface{}{},
			},
			want: map[string]interface{}{
				"hclPercentChange":    "-36.36",
				"absHclPercentChange": "36.36",
			},
		},
		{
			name: "test sets tf_var marker",
			args: args{
				r:    output.Root{},
				hclR: output.Root{},
				osVars: map[string]string{
					"TF_VAR_TEST": "testing",
				},
				env: map[string]interface{}{},
			},
			want: map[string]interface{}{
				"hclPercentChange":    "0.00",
				"absHclPercentChange": "0.00",
				"tfVarPresent":        true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.osVars != nil {
				for k, v := range tt.args.osVars {
					os.Setenv(k, v)
				}

				defer func() {
					for k, _ := range tt.args.osVars {
						os.Unsetenv(k)
					}
				}()
			}

			main.AddHCLEnvVars(tt.args.r, tt.args.hclR, tt.args.env)
			assert.Equal(t, tt.want, tt.args.env)
		})
	}
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}
