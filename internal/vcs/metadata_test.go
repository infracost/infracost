package vcs

import (
	"reflect"
	"testing"
)

func Test_urlStringToRemote(t *testing.T) {
	tests := []struct {
		input string
		want  Remote
	}{
		{"git@github.com:org/repo.git", Remote{
			Host: "github.com",
			URL:  "https://github.com/org/repo.git",
		}},
		{"https://github.com/org/repo.git", Remote{
			Host: "github.com",
			URL:  "https://github.com/org/repo.git",
		}},
		{"git@gitlab.com:org/repo.git", Remote{
			Host: "gitlab.com",
			URL:  "https://gitlab.com/org/repo.git",
		}},
		{"https://gitlab.com/org/repo.git", Remote{
			Host: "gitlab.com",
			URL:  "https://gitlab.com/org/repo.git",
		}},
		{"git@bitbucket.org:org/repo.git", Remote{
			Host: "bitbucket.org",
			URL:  "https://bitbucket.org/org/repo.git",
		}},
		{"https://user@bitbucket.org/org/repo.git", Remote{
			Host: "bitbucket.org",
			URL:  "https://bitbucket.org/org/repo.git",
		}},
		{"https://user@dev.azure.com/org/project/_git/repo", Remote{
			Host: "dev.azure.com",
			URL:  "https://dev.azure.com/org/project/_git/repo",
		}},
		{"git@ssh.dev.azure.com:v3/org/project/repo", Remote{
			Host: "dev.azure.com",
			URL:  "https://dev.azure.com/org/project/_git/repo",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := urlStringToRemote(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("urlStringToRemote() = %v, want %v", got, tt.want)
			}
		})
	}
}
