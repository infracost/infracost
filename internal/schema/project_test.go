package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNameFromRepoURL(t *testing.T) {
	tests := []struct {
		repoURL string
		name    string
	}{
		{"git@github.com:org/repo.git", "org/repo"},
		{"https://github.com/org/repo.git", "org/repo"},
		{"git@gitlab.com:org/repo.git", "org/repo"},
		{"https://gitlab.com/org/repo.git", "org/repo"},
		{"git@bitbucket.org:org/repo.git", "org/repo"},
		{"https://user@bitbucket.org/org/repo.git", "org/repo"},
		{"https://user@dev.azure.com/org/project/_git/repo", "org/project/repo"},
		{"git@ssh.dev.azure.com:v3/org/project/repo", "org/project/repo"},
	}

	for _, test := range tests {
		actual := nameFromRepoURL(test.repoURL)
		assert.Equal(t, test.name, actual)
	}
}

func TestGenerateProjectName(t *testing.T) {
	type args struct {
		metadata         *ProjectMetadata
		repoURL          string
		dashboardEnabled bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "azure repo url should show org/project/repo with spaces",
			args: args{
				metadata: &ProjectMetadata{},
				repoURL:  "https://infracost-user@dev.azure.com/infracost/my%20project/_git/my%20repo",
			},
			want: "infracost/my project/my repo",
		},
		{
			name: "github repo https url",
			args: args{
				metadata: &ProjectMetadata{},
				repoURL:  "https://github.com/infracost/infracost.git",
			},
			want: "infracost/infracost",
		},
		{
			name: "github repo without '.git' extension",
			args: args{
				metadata: &ProjectMetadata{},
				repoURL:  "https://github.com/infracost/infracost",
			},
			want: "infracost/infracost",
		},
		{
			name: "github repo with domain in repo name",
			args: args{
				metadata: &ProjectMetadata{},
				repoURL:  "https://github.com/infracost.io/infracost.git",
			},
			want: "infracost.io/infracost",
		},
		{
			name: "github repo ssh url",
			args: args{
				metadata: &ProjectMetadata{},
				repoURL:  "git@github.com:infracost/infracost.git",
			},
			want: "infracost/infracost",
		},
		{
			name: "gitlab repo https url",
			args: args{
				metadata: &ProjectMetadata{},
				repoURL:  "https://gitlab.com/infracost/infracost-gitlab-ci.git",
			},
			want: "infracost/infracost-gitlab-ci",
		},
		{
			name: "gitlab repo ssh url",
			args: args{
				metadata: &ProjectMetadata{},
				repoURL:  "git@gitlab.com:infracost/infracost-gitlab-ci.git",
			},
			want: "infracost/infracost-gitlab-ci",
		},
		{
			name: "bitbucket repo https url",
			args: args{
				metadata: &ProjectMetadata{},
				repoURL:  "https://hugorutinfracost@bitbucket.org/infracost/infracost-bitbucket-pipeline.git",
			},
			want: "infracost/infracost-bitbucket-pipeline",
		},
		{
			name: "bitbucket repo ssh url",
			args: args{
				metadata: &ProjectMetadata{},
				repoURL:  "git@bitbucket.org:infracost/infracost-bitbucket-pipeline.git",
			},
			want: "infracost/infracost-bitbucket-pipeline",
		},
		{
			name: "bitbucket repo ssh url with port",
			args: args{
				metadata: &ProjectMetadata{},
				repoURL:  "git@bitbucket.org:8888/~test.infracost.io/infracost-bitbucket-pipeline.git",
			},
			want: "~test.infracost.io/infracost-bitbucket-pipeline",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.args.metadata.GenerateProjectName(tt.args.repoURL, tt.args.dashboardEnabled), "GenerateProjectName(%v, %v)", tt.args.repoURL, tt.args.dashboardEnabled)
		})
	}
}
