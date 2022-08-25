package vcs

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
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

func Test_metadataFetcher_GetLocalMetadata(t *testing.T) {
	tmp := t.TempDir()
	obj := createLocalRepoWithCommit(t, tmp)
	t.Setenv("GITHUB_ACTIONS", "")

	test := false
	m := metadataFetcher{
		mu:     &keyMutex{},
		client: &http.Client{Timeout: time.Second * 5},
		test:   &test,
	}

	actual, err := m.Get(tmp)
	assert.NoError(t, err)

	assert.Equal(t, Metadata{
		Remote: Remote{
			Host: "github.com",
			URL:  "https://github.com/git-fixtures/basic.git",
		},
		Branch: Branch{
			Name: "master",
		},
		Commit: Commit{
			SHA:         obj.Hash.String(),
			AuthorName:  obj.Author.Name,
			AuthorEmail: obj.Author.Email,
			Time:        obj.Author.When,
			Message:     obj.Message,
		},
		PullRequest: nil,
		Pipeline:    nil,
	}, actual)
}

func Test_metadataFetcher_GetLocalMetadataMergesWithEnv(t *testing.T) {
	tmp := t.TempDir()
	obj := createLocalRepoWithCommit(t, tmp)
	providedName := "test provided name"

	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("INFRACOST_COMMIT_AUTHOR_NAME", providedName)

	pullID := "1234"
	t.Setenv("INFRACOST_VCS_PULL_REQUEST_ID", pullID)

	test := false
	m := metadataFetcher{
		mu:     &keyMutex{},
		client: &http.Client{Timeout: time.Second * 5},
		test:   &test,
	}

	actual, err := m.Get(tmp)
	assert.NoError(t, err)

	assert.Equal(t, Metadata{
		Remote: Remote{
			Host: "github.com",
			URL:  "https://github.com/git-fixtures/basic.git",
		},
		Branch: Branch{
			Name: "master",
		},
		Commit: Commit{
			SHA:         obj.Hash.String(),
			AuthorName:  providedName,
			AuthorEmail: obj.Author.Email,
			Time:        obj.Author.When,
			Message:     obj.Message,
		},
		PullRequest: &PullRequest{ID: pullID},
		Pipeline:    nil,
	}, actual)
}

func createLocalRepoWithCommit(t *testing.T, tmp string) *object.Commit {
	t.Helper()
	r, err := git.PlainInit(tmp, false)
	assert.NoError(t, err)

	_, err = r.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"https://github.com/git-fixtures/basic.git"},
	})
	assert.NoError(t, err)

	w, err := r.Worktree()
	assert.NoError(t, err)

	filename := filepath.Join(tmp, "example-git-file")
	err = ioutil.WriteFile(filename, []byte("hello world!"), 0600)
	assert.NoError(t, err)

	commit, err := w.Commit("example go-git commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)

	obj, err := r.CommitObject(commit)
	assert.NoError(t, err)

	return obj
}

func Test_metadataFetcher_GetReturnsUserDefinedEnvs(t *testing.T) {
	tt := time.Unix(1661350126, 0).UTC()

	t.Setenv("INFRACOST_VCS_PULL_REQUEST_ID", "INFRACOST_VCS_PULL_REQUEST_ID_VALUE")
	t.Setenv("INFRACOST_VCS_PROVIDER", "github")
	t.Setenv("INFRACOST_VCS_PULL_REQUEST_TITLE", "INFRACOST_VCS_PULL_REQUEST_TITLE_VALUE")
	t.Setenv("INFRACOST_VCS_PULL_REQUEST_AUTHOR", "INFRACOST_VCS_PULL_REQUEST_AUTHOR_VALUE")
	t.Setenv("INFRACOST_VCS_BRANCH", "INFRACOST_BRANCH_VALUE")
	t.Setenv("INFRACOST_VCS_BASE_BRANCH", "INFRACOST_VCS_BASE_BRANCH_VALUE")
	t.Setenv("INFRACOST_VCS_PULL_REQUEST_URL", "https://github.com/infracost/test-repo/pull/1979")
	t.Setenv("INFRACOST_COMMIT", "INFRACOST_COMMIT_VALUE")
	t.Setenv("INFRACOST_COMMIT_AUTHOR_NAME", "INFRACOST_COMMIT_AUTHOR_NAME_VALUE")
	t.Setenv("INFRACOST_COMMIT_AUTHOR_EMAIL", "INFRACOST_COMMIT_AUTHOR_EMAIL_VALUE")
	timestamp := fmt.Sprintf("%d", tt.Unix())
	t.Setenv("INFRACOST_COMMIT_TIMESTAMP", timestamp)
	t.Setenv("INFRACOST_COMMIT_MESSAGE", "INFRACOST_COMMIT_MESSAGE_VALUE")
	t.Setenv("INFRACOST_VCS_REPOSITORY_URL", "https://github.com/infracost/test-repo.git")
	t.Setenv("INFRACOST_BRANCH", "INFRACOST_BRANCH_VALUE")
	t.Setenv("INFRACOST_VCS_PIPELINE_RUN_ID", "INFRACOST_VCS_PIPELINE_RUN_ID_VALUE")

	test := false
	m := metadataFetcher{
		mu:     &keyMutex{},
		client: &http.Client{Timeout: time.Second * 5},
		test:   &test,
	}

	actual, _ := m.Get(t.TempDir())

	assert.Equal(t, Metadata{
		Remote: Remote{
			Host: "github.com",
			URL:  "https://github.com/infracost/test-repo.git",
		},
		Branch: Branch{
			Name: "INFRACOST_BRANCH_VALUE",
		},
		Commit: Commit{
			SHA:         "INFRACOST_COMMIT_VALUE",
			AuthorName:  "INFRACOST_COMMIT_AUTHOR_NAME_VALUE",
			AuthorEmail: "INFRACOST_COMMIT_AUTHOR_EMAIL_VALUE",
			Time:        tt,
			Message:     "INFRACOST_COMMIT_MESSAGE_VALUE",
		},
		PullRequest: &PullRequest{
			ID:           "INFRACOST_VCS_PULL_REQUEST_ID_VALUE",
			VCSProvider:  "github",
			Title:        "INFRACOST_VCS_PULL_REQUEST_TITLE_VALUE",
			Author:       "INFRACOST_VCS_PULL_REQUEST_AUTHOR_VALUE",
			SourceBranch: "INFRACOST_BRANCH_VALUE",
			BaseBranch:   "INFRACOST_VCS_BASE_BRANCH_VALUE",
			URL:          "https://github.com/infracost/test-repo/pull/1979",
		},
		Pipeline: &Pipeline{ID: "INFRACOST_VCS_PIPELINE_RUN_ID_VALUE"},
	}, actual)
}
