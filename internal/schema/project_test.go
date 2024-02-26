package schema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/infracost/infracost/internal/vcs"
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

func TestDetectProjectMetadataVCSSubPath(t *testing.T) {
	type args struct {
		wdir     string
		expected string
	}

	type test struct {
		name          string
		createProject func(t *testing.T) (input args, cleanup func())
	}

	tests := []test{
		{
			name: "sub path detected correctly",
			createProject: func(t *testing.T) (input args, cleanup func()) {
				t.Setenv("INFRACOST_VCS_REPOSITORY_URL", "test")
				// set the TMPDDIR variable to the non testing directory so that we can accurately replicate
				// known bugs which set /private prefix in Darwin for /tmp
				t.Setenv("TMPDIR", "")

				tmp, clean := tmpDir(t, os.TempDir())
				_, err := git.PlainInit(tmp, false)
				assert.NoError(t, err)

				sub, _ := tmpDir(t, tmp)
				child, err := filepath.Rel(tmp, sub)
				assert.NoError(t, err)

				return args{
					wdir:     sub,
					expected: child,
				}, clean
			},
		},
		{
			name: "root github project returns empty string",
			createProject: func(t *testing.T) (input args, cleanup func()) {
				t.Setenv("INFRACOST_VCS_REPOSITORY_URL", "test")
				// set the TMPDDIR variable to the non testing directory so that we can accurately replicate
				// known bugs which set /private prefix in Darwin for /tmp
				t.Setenv("TMPDIR", "")

				tmp, clean := tmpDir(t, os.TempDir())
				_, err := git.PlainInit(tmp, false)
				assert.NoError(t, err)

				return args{
					wdir:     tmp,
					expected: "",
				}, clean
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, cleanup := tt.createProject(t)
			defer cleanup()

			meta := DetectProjectMetadata(input.wdir)
			assert.Equal(t, input.expected, meta.VCSSubPath)
		})
	}

}

func tmpDir(t *testing.T, path string) (string, func()) {
	t.Helper()

	fs := osfs.New(path)
	path, err := util.TempDir(fs, "", "")
	require.NoError(t, err)

	return fs.Join(fs.Root(), path), func() {
		t.Helper()

		err := util.RemoveAll(fs, path)
		assert.NoError(t, err)
	}
}
