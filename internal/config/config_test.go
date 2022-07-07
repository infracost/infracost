package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigLoadFromConfigFile(t *testing.T) {
	tmp := t.TempDir()
	tests := []struct {
		name     string
		contents []byte
		expected []*Project
		error    error
	}{
		{
			name: "should parse valid projects",
			contents: []byte(`version: 0.1

projects:
  - path: path/to/my_terraform
  - path: path/to/my_terraform_two
    terraform_plan_flags: "-var-file=prod.tfvars -var-file=us-east.tfvars"
    terraform_binary: "~/bin/terraform"
    terraform_workspace: "development"
    terraform_cloud_host: "cloud_host"
    terraform_cloud_token: "cloud_token"
    usage_file: "usage/file"
    terraform_use_state: true
`),
			expected: []*Project{
				{
					Path: "path/to/my_terraform",
				},
				{
					Path:                "path/to/my_terraform_two",
					TerraformPlanFlags:  "-var-file=prod.tfvars -var-file=us-east.tfvars",
					TerraformBinary:     "~/bin/terraform",
					TerraformWorkspace:  "development",
					TerraformCloudHost:  "cloud_host",
					TerraformCloudToken: "cloud_token",
					UsageFile:           "usage/file",
					TerraformUseState:   true,
				},
			},
		},
		{
			name: "should return error if no projects given",
			contents: []byte(`version: 0.1

projects:
`),
			error: &YamlError{raw: ErrorNilProjects},
		},
		{
			name: "should return panic error wrapped with invalid config file error",
			contents: []byte(`version: 0.1

projects:
  - afdas: safasfddas
		`),
			error: fmt.Errorf("%w: yaml: line 5: found a tab character that violates indentation", ErrorInvalidConfigFile),
		},
		{
			name: "should error invalid project key given",
			contents: []byte(`version: 0.1

projects:
  - path: path/to/my_terraform
    invalid_key: "test"
`),
			error: &YamlError{
				base: "config file is invalid, see https://infracost.io/config-file for valid options",
				errors: []error{
					&YamlError{
						base: "project config defined for path: [path/to/my_terraform] is invalid",
						errors: []error{
							errors.New("invalid_key is not a valid project configuration option"),
						},
					},
				},
			},
		},
		{
			name: "should error invalid version given",
			contents: []byte(`version: 81923.1

projects:
  - path: path/to/my_terraform
`),
			error: &YamlError{
				base: "config file is invalid, see https://infracost.io/config-file for file specification",
				errors: []error{
					errors.New("version '81923.1' is not supported, valid versions are 0.1 ≤ x ≤ 0.1"),
				},
			},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Config{}
			path := filepath.Join(tmp, fmt.Sprintf("conf-%d.yaml", i))
			err := os.WriteFile(path, tt.contents, os.ModePerm)
			require.NoError(t, err)

			// we need to remove INFRACOST_TERRAFORM_CLOUD_TOKEN value for these tests.
			// as CI uses INFRACOST_TERRAFORM_CLOUD_TOKEN for private registry tests. This means the expected value
			// will be inconsistent and show "***".
			key := "INFRACOST_TERRAFORM_CLOUD_TOKEN"
			v := os.Getenv(key)
			os.Unsetenv(key)

			if v != "" {
				defer func() {
					os.Setenv(key, v)
				}()
			}

			err = c.LoadFromConfigFile(path)

			require.Equal(t, tt.error, err)
			require.EqualValues(t, tt.expected, c.Projects)
		})
	}
}
