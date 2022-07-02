package main_test

import (
	"os"
	"testing"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	main "github.com/infracost/infracost/cmd/infracost"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/schema"
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

func TestAddHCLEnvVars(t *testing.T) {
	type args struct {
		r           output.Root
		hclR        output.Root
		osVars      map[string]string
		env         map[string]interface{}
		pctx        []*config.ProjectContext
		projects    []*schema.Project
		hclProjects []*schema.Project
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
				"hclRunTimeMs":        int64(0),
				"tfRunTimeMs":         int64(0),
				"hclMissingResources": []string{},
				"hclResourceDiff":     map[string][]string{},
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
				"hclRunTimeMs":        int64(0),
				"tfRunTimeMs":         int64(0),
				"hclMissingResources": []string{},
				"hclResourceDiff":     map[string][]string{},
			},
		},
		{
			name: "test sums time ms for projects",
			args: args{
				pctx: []*config.ProjectContext{
					newProjectContextWithCtx(map[string]interface{}{
						"hclProjectRunTimeMs": int64(10),
						"tfProjectRunTimeMs":  int64(20),
					}),
					newProjectContextWithCtx(map[string]interface{}{
						"hclProjectRunTimeMs": int64(50),
						"tfProjectRunTimeMs":  int64(70),
					}),
				},
				r:    output.Root{},
				hclR: output.Root{},
				env:  map[string]interface{}{},
			},
			want: map[string]interface{}{
				"hclPercentChange":    "0.00",
				"absHclPercentChange": "0.00",
				"hclRunTimeMs":        int64(60),
				"tfRunTimeMs":         int64(90),
				"hclMissingResources": []string{},
				"hclResourceDiff":     map[string][]string{},
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
				"hclRunTimeMs":        int64(0),
				"tfRunTimeMs":         int64(0),
				"hclMissingResources": []string{},
				"hclResourceDiff":     map[string][]string{},
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
				"hclRunTimeMs":        int64(0),
				"tfRunTimeMs":         int64(0),
				"hclMissingResources": []string{},
				"hclResourceDiff":     map[string][]string{},
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
				"hclRunTimeMs":        int64(0),
				"tfRunTimeMs":         int64(0),
				"hclMissingResources": []string{},
				"hclResourceDiff":     map[string][]string{},
			},
		},
		{
			name: "test collects missing HCL resources",
			args: args{
				r:    output.Root{},
				hclR: output.Root{},
				env:  map[string]interface{}{},
				projects: []*schema.Project{
					{
						Name: "test-project",
						Resources: []*schema.Resource{
							{
								Name:         "aws_instance.my_app",
								ResourceType: "aws_instance",
								MonthlyCost:  decimalPtr(decimal.NewFromInt(100)),
							},
							{
								Name:         "module.vpc.aws_subnet.public[0]",
								ResourceType: "aws_subnet",
								MonthlyCost:  nil,
							},
							{
								Name:         "module.vpc.aws_subnet.public[1]",
								ResourceType: "aws_subnet",
								MonthlyCost:  nil,
							},
						},
					},
				},
				hclProjects: []*schema.Project{
					{
						Name: "test-project",
						Resources: []*schema.Resource{
							{
								Name:         "aws_instance.my_app",
								ResourceType: "aws_instance",
								MonthlyCost:  decimalPtr(decimal.NewFromInt(100)),
							},
						},
					},
				},
			},
			want: map[string]interface{}{
				"hclPercentChange":    "0.00",
				"absHclPercentChange": "0.00",
				"hclRunTimeMs":        int64(0),
				"tfRunTimeMs":         int64(0),
				"hclMissingResources": []string{"aws_subnet"},
				"hclResourceDiff":     map[string][]string{},
			},
		},
		{
			name: "test collects HCL resources' cost discrepancies",
			args: args{
				r:    output.Root{},
				hclR: output.Root{},
				env:  map[string]interface{}{},
				projects: []*schema.Project{
					{
						Name: "test-project",
						Resources: []*schema.Resource{
							{
								Name:         "module.vpc.aws_subnet.public",
								ResourceType: "aws_subnet",
								MonthlyCost:  decimalPtr(decimal.NewFromInt(10)),
							},
							{
								Name:         "aws_instance.my_app",
								ResourceType: "aws_instance",
								MonthlyCost:  decimalPtr(decimal.NewFromInt(100)),
							},
							{
								Name:         "aws_instance.another_app",
								ResourceType: "aws_instance",
								MonthlyCost:  decimalPtr(decimal.NewFromFloat(80)),
							},
							{
								Name:         "eks.aws_eks_cluster.this",
								ResourceType: "aws_eks_cluster",
								MonthlyCost:  decimalPtr(decimal.NewFromInt(10)),
							},
							{
								Name:         "aws_db_instance.test_db",
								ResourceType: "aws_db_instance",
								MonthlyCost:  nil,
							},
						},
					},
				},
				hclProjects: []*schema.Project{
					{
						Name: "test-project",
						Resources: []*schema.Resource{
							{
								Name:         "module.vpc.aws_subnet.public",
								ResourceType: "aws_subnet",
								MonthlyCost:  decimalPtr(decimal.NewFromInt(10)),
							},
							{
								Name:         "aws_instance.my_app",
								ResourceType: "aws_instance",
								MonthlyCost:  decimalPtr(decimal.NewFromInt(50)),
							},
							{
								Name:         "aws_instance.another_app",
								ResourceType: "aws_instance",
								MonthlyCost:  decimalPtr(decimal.NewFromInt(150)),
							},
							{
								Name:         "eks.aws_eks_cluster.this",
								ResourceType: "aws_eks_cluster",
								MonthlyCost:  nil,
							},
							{
								Name:         "aws_db_instance.test_db",
								ResourceType: "aws_db_instance",
								MonthlyCost:  decimalPtr(decimal.NewFromInt(10)),
							},
						},
					},
				},
			},
			want: map[string]interface{}{
				"hclPercentChange":    "0.00",
				"absHclPercentChange": "0.00",
				"hclRunTimeMs":        int64(0),
				"tfRunTimeMs":         int64(0),
				"hclMissingResources": []string{},
				"hclResourceDiff": map[string][]string{
					"aws_instance":    {"-50.00", "87.50"},
					"aws_eks_cluster": {"-100.00"},
					"aws_db_instance": {"100.00"},
				},
			},
		},
		{
			name: "test skips project when names mismatch",
			args: args{
				r:    output.Root{},
				hclR: output.Root{},
				env:  map[string]interface{}{},
				projects: []*schema.Project{
					{
						Name: "test-project",
						Resources: []*schema.Resource{
							{
								Name:         "aws_instance.my_app",
								ResourceType: "aws_instance",
								MonthlyCost:  decimalPtr(decimal.NewFromInt(100)),
							},
						},
					},
				},
				hclProjects: []*schema.Project{
					{
						Name: "hcl-test-project",
						Resources: []*schema.Resource{
							{
								Name:         "aws_instance.my_app",
								ResourceType: "aws_instance",
								MonthlyCost:  decimalPtr(decimal.NewFromInt(50)),
							},
						},
					},
				},
			},
			want: map[string]interface{}{
				"hclPercentChange":    "0.00",
				"absHclPercentChange": "0.00",
				"hclRunTimeMs":        int64(0),
				"tfRunTimeMs":         int64(0),
				"hclMissingResources": []string{},
				"hclResourceDiff":     map[string][]string{},
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
					for k := range tt.args.osVars {
						os.Unsetenv(k)
					}
				}()
			}

			main.AddHCLEnvVars(tt.args.pctx, tt.args.r, tt.args.projects, tt.args.hclR, tt.args.hclProjects, tt.args.env)
			assert.Equal(t, tt.want, tt.args.env)
		})
	}
}

func newProjectContextWithCtx(m map[string]interface{}) *config.ProjectContext {
	ctx := config.NewProjectContext(nil, nil, log.Fields{})

	for k, v := range m {
		ctx.SetContextValue(k, v)
	}

	return ctx
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}
