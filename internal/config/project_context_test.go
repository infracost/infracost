package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
