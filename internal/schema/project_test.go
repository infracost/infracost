package schema

import (
	"strings"
	"testing"

	"github.com/infracost/infracost/internal/vcs"
	"github.com/stretchr/testify/assert"
)

func TestGenerateProjectName(t *testing.T) {
	type args struct {
		metadata         *ProjectMetadata
		remote           vcs.Remote
		dashboardEnabled bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "repo with remote metadata",
			args: args{
				metadata: &ProjectMetadata{Path: "path/to/repo/test-repo"},
				remote:   vcs.Remote{Name: "infracost/remote-repo"},
			},
			want: "infracost/remote-repo",
		},
		{
			name: "repo with remote metadata and VCS subpath",
			args: args{
				metadata: &ProjectMetadata{Path: "path/to/repo/test-repo", VCSSubPath: "sub/path"},
				remote:   vcs.Remote{Name: "infracost/remote-repo"},
			},
			want: "infracost/remote-repo/sub/path",
		},
		{
			name: "repo without empty remote data",
			args: args{
				metadata: &ProjectMetadata{Path: "path/to/repo/test-repo"},
				remote:   vcs.Remote{},
			},
			want: "path/to/repo/test-repo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(
				t,
				tt.want,
				tt.args.metadata.GenerateProjectName(tt.args.remote, tt.args.dashboardEnabled),
				"GenerateProjectName(%v, %v)",
				tt.args.remote,
				tt.args.dashboardEnabled,
			)
		})
	}

	t.Run("repo without remote metadata, Infracost Cloud is enabled", func(t *testing.T) {
		testArgs := args{
			metadata:         &ProjectMetadata{Path: "path/to/repo/test-repo"},
			remote:           vcs.Remote{},
			dashboardEnabled: true,
		}

		result := testArgs.metadata.GenerateProjectName(testArgs.remote, testArgs.dashboardEnabled)
		assert.Len(t, result, len("project_xxxxxxxx"))
		assert.True(t, strings.HasPrefix(result, "project_"))
	})
}
