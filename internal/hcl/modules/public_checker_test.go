package modules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsPublic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	tests := []struct {
		name       string
		moduleAddr string
		expected   bool
	}{
		{
			name:       "public module",
			moduleAddr: "https://github.com/terraform-aws-modules/terraform-aws-alb?ref=46852b88a2bf09bd097e6ad3d1acc9a763cf9005",
			expected:   true,
		},
		{
			name:       "private module",
			moduleAddr: "https://github.com/infracost/infracost-modules?ref=0.0.1",
			expected:   false,
		},
		{
			name:       "public module from git",
			moduleAddr: "git::https://github.com/terraform-aws-modules/terraform-aws-alb?ref=46852b88a2bf09bd097e6ad3d1acc9a763cf9005",
			expected:   true,
		},
		{
			name:       "private module from git",
			moduleAddr: "git::https://github.com/infracost/infracost-modules?ref=0.0.1",
			expected:   false,
		},
		{
			name:       "public with no scheme",
			moduleAddr: "github.com/infracost/infracost-modules?ref=0.0.1",
			expected:   false,
		},
		{
			name:       "public with git",
			moduleAddr: "git@github.com:terraform-aws-modules/terraform-aws-alb?ref=46852b88a2bf09bd097e6ad3d1acc9a763cf9005",
			expected:   false,
		},
		{
			name:       "with username and password",
			moduleAddr: "https://username:password@jfrog.infraacost.com",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewHttpPublicModuleChecker()
			got, err := checker.IsPublicModule(tt.moduleAddr)
			require.NoError(t, err)
			require.Equal(t, tt.expected, got)
		})
	}

}
