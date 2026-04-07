package vcs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/infracost/infracost/internal/sync"
)

func Test_urlStringToRemote(t *testing.T) {
	tests := []struct {
		input string
		want  Remote
	}{
		{"git@github.com:org/repo.git", Remote{
			Host: "github.com",
			Name: "org/repo",
			URL:  "https://github.com/org/repo.git",
		}},
		{"https://github.com/org/repo.git", Remote{
			Host: "github.com",
			Name: "org/repo",
			URL:  "https://github.com/org/repo.git",
		}},
		{"git@gitlab.com:org/repo.git", Remote{
			Host: "gitlab.com",
			Name: "org/repo",
			URL:  "https://gitlab.com/org/repo.git",
		}},
		{"https://gitlab.com/org/repo.git", Remote{
			Host: "gitlab.com",
			Name: "org/repo",
			URL:  "https://gitlab.com/org/repo.git",
		}},
		{"git@bitbucket.org:org/repo.git", Remote{
			Host: "bitbucket.org",
			Name: "org/repo",
			URL:  "https://bitbucket.org/org/repo.git",
		}},
		{"https://user@bitbucket.org/org/repo.git", Remote{
			Host: "bitbucket.org",
			Name: "org/repo",
			URL:  "https://bitbucket.org/org/repo.git",
		}},
		{"https://user@dev.azure.com/org/project/_git/repo", Remote{
			Host: "dev.azure.com",
			Name: "org/project/repo",
			URL:  "https://dev.azure.com/org/project/_git/repo",
		}},
		{"git@ssh.dev.azure.com:v3/org/project/repo", Remote{
			Host: "dev.azure.com",
			Name: "org/project/repo",
			URL:  "https://dev.azure.com/org/project/_git/repo",
		}},
		{"https://user@bitbucket.custom.org:8888/custom.org/repo", Remote{
			Host: "bitbucket.custom.org",
			Name: "custom.org/repo",
			URL:  "https://bitbucket.custom.org:8888/custom.org/repo",
		}},
		{"git@bitbucket.org:8888/~test.infracost.io/infracost-bitbucket-pipeline.git", Remote{
			Host: "bitbucket.org",
			Name: "~test.infracost.io/infracost-bitbucket-pipeline",
			URL:  "https://bitbucket.org:8888/~test.infracost.io/infracost-bitbucket-pipeline.git",
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
func Test_metadataFetcher_GetAzureMetadata(t *testing.T) {
	tmp := t.TempDir()
	mux := &http.ServeMux{}
	mux.HandleFunc("/_apis/git/repositories/myrepo/pullRequests/456", func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		assert.Equal(t, "Basic YXpkbzp0ZXN0LXRva2Vu", r.Header.Get("Authorization"))
		_, err := w.Write([]byte(`{
			"title": "pr title",
			"createdBy": { "uniqueName": "test-user" }
		}`))
		assert.NoError(t, err)
	})

	s := httptest.NewServer(mux)
	defer s.Close()

	_, lastCommit := createLocalRepoWithCommits(t, tmp)
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("BUILD_REPOSITORY_PROVIDER", "tfsgit")
	t.Setenv("BUILD_REPOSITORY_URI", "https://user@dev.azure.com/org/project/_git/myrepo")
	t.Setenv("BUILD_BUILDID", "123")
	t.Setenv("SYSTEM_PULLREQUEST_PULLREQUESTID", "456")
	t.Setenv("SYSTEM_PULLREQUEST_SOURCEREPOSITORYURI", "https://user@dev.azure.com/org/project/_git/myrepo")
	t.Setenv("SYSTEM_PULLREQUEST_SOURCEBRANCH", "test")
	t.Setenv("SYSTEM_PULLREQUEST_TARGETBRANCH", "main")
	t.Setenv("SYSTEM_ACCESSTOKEN", "test-token")
	t.Setenv("SYSTEM_COLLECTIONURI", fmt.Sprintf("%s/", s.URL))
	t.Setenv("BUILD_REPOSITORY_ID", "myrepo")

	test := false
	m := metadataFetcher{
		mu:     &sync.KeyMutex{},
		client: &http.Client{Timeout: time.Second * 5},
		test:   &test,
	}

	actual, err := m.Get(tmp, nil)
	assert.NoError(t, err)
	assert.Equal(t, Metadata{
		Remote: Remote{
			Host: "dev.azure.com",
			Name: "org/project/myrepo",
			URL:  "https://dev.azure.com/org/project/_git/myrepo",
		},
		Branch: Branch{
			Name: "master",
		},
		Commit: Commit{
			SHA:         lastCommit.Hash.String(),
			AuthorName:  lastCommit.Author.Name,
			AuthorEmail: lastCommit.Author.Email,
			Time:        lastCommit.Author.When,
			Message:     lastCommit.Message,
			ChangedObjects: []string{
				filepath.Join(tmp, "added-file"),
			},
		},
		PullRequest: &PullRequest{
			ID:           "456",
			Title:        "pr title",
			Author:       "test-user",
			VCSProvider:  "azure_devops_tfsgit",
			SourceBranch: "test",
			BaseBranch:   "main",
			URL:          "https://dev.azure.com/org/project/_git/myrepo/pullrequest/456",
		},
		Pipeline: &Pipeline{ID: "123"},
	}, actual)
}

func Test_metadataFetcher_GetGitlabMetadata(t *testing.T) {
	tmp := t.TempDir()
	mux := &http.ServeMux{}
	token := "dummytoken"
	username := "apiusername"
	mux.HandleFunc("/api/v4/projects/123/merge_requests/54", func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		assert.Equal(t, token, r.Header.Get("Private-Token"))
		_, err := w.Write(fmt.Appendf(nil, `{
			"author": {
				"username": %q
			}
		}`, username))
		assert.NoError(t, err)
	})

	s := httptest.NewServer(mux)
	defer s.Close()

	_, lastCommit := createLocalRepoWithCommits(t, tmp)
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("GITLAB_TOKEN", token)
	t.Setenv("CI_API_V4_URL", s.URL+"/api/v4")
	t.Setenv("CI_PROJECT_ID", "123")
	t.Setenv("CI_MERGE_REQUEST_IID", "54")
	t.Setenv("GITLAB_CI", "true")
	t.Setenv("CI_PROJECT_URL", "https://gitlab.com/my-project")
	t.Setenv("CI_MERGE_REQUEST_TITLE", "pr title")
	t.Setenv("CI_MERGE_REQUEST_SOURCE_BRANCH_NAME", "test")
	t.Setenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME", "main")
	t.Setenv("CI_PIPELINE_ID", "123")

	test := false
	m := metadataFetcher{
		mu:     &sync.KeyMutex{},
		client: &http.Client{Timeout: time.Second * 5},
		test:   &test,
	}

	actual, err := m.Get(tmp, nil)
	assert.NoError(t, err)
	assert.Equal(t, Metadata{
		Remote: Remote{
			Host: "gitlab.com",
			Name: "my-project",
			URL:  "https://gitlab.com/my-project",
		},
		Branch: Branch{
			Name: "master",
		},
		Commit: Commit{
			SHA:         lastCommit.Hash.String(),
			AuthorName:  lastCommit.Author.Name,
			AuthorEmail: lastCommit.Author.Email,
			Time:        lastCommit.Author.When,
			Message:     lastCommit.Message,
			ChangedObjects: []string{
				filepath.Join(tmp, "added-file"),
			},
		},
		PullRequest: &PullRequest{
			ID:           "54",
			Title:        "pr title",
			Author:       username,
			VCSProvider:  "gitlab",
			SourceBranch: "test",
			BaseBranch:   "main",
			URL:          "https://gitlab.com/my-project/merge_requests/54",
		},
		Pipeline: &Pipeline{ID: "123"},
	}, actual)
}

func Test_metadataFetcher_GetLocalMetadata(t *testing.T) {
	tmp := t.TempDir()
	_, lastCommit := createLocalRepoWithCommits(t, tmp)
	t.Setenv("GITHUB_ACTIONS", "")

	test := false
	m := metadataFetcher{
		mu:     &sync.KeyMutex{},
		client: &http.Client{Timeout: time.Second * 5},
		test:   &test,
	}

	actual, err := m.Get(tmp, nil)
	assert.NoError(t, err)

	assert.Equal(t, Metadata{
		Remote: Remote{
			Host: "github.com",
			Name: "git-fixtures/basic",
			URL:  "https://github.com/git-fixtures/basic.git",
		},
		Branch: Branch{
			Name: "master",
		},
		Commit: Commit{
			SHA:         lastCommit.Hash.String(),
			AuthorName:  lastCommit.Author.Name,
			AuthorEmail: lastCommit.Author.Email,
			Time:        lastCommit.Author.When,
			Message:     lastCommit.Message,
			ChangedObjects: []string{
				filepath.Join(tmp, "added-file"),
			},
		},
		PullRequest: nil,
		Pipeline:    nil,
	}, actual)
}

func Test_metadataFetcher_GetLocalMetadata_WithGitDiffTarget(t *testing.T) {
	tmp := t.TempDir()
	r, _ := createLocalRepoWithCommits(t, tmp)

	w, err := r.Worktree()
	require.NoError(t, err)
	err = w.Checkout(&git.CheckoutOptions{
		Create: true,
		Branch: plumbing.NewBranchReferenceName("a-branch"),
	})
	require.NoError(t, err)

	lastCommit := createCommit(t, r, w, tmp, "branch-file", "something")

	t.Setenv("GITHUB_ACTIONS", "")
	base := "master"

	test := false
	m := metadataFetcher{
		mu:     &sync.KeyMutex{},
		client: &http.Client{Timeout: time.Second * 5},
		test:   &test,
	}

	actual, err := m.Get(tmp, &base)
	assert.NoError(t, err)

	assert.Equal(t, Metadata{
		Remote: Remote{
			Host: "github.com",
			Name: "git-fixtures/basic",
			URL:  "https://github.com/git-fixtures/basic.git",
		},
		Branch: Branch{
			Name: "a-branch",
		},
		Commit: Commit{
			SHA:         lastCommit.Hash.String(),
			AuthorName:  lastCommit.Author.Name,
			AuthorEmail: lastCommit.Author.Email,
			Time:        lastCommit.Author.When,
			Message:     lastCommit.Message,
			ChangedObjects: []string{
				filepath.Join(tmp, "branch-file"),
			},
			GitDiffTarget: &base,
		},
		PullRequest: nil,
		Pipeline:    nil,
	}, actual)
}

func Test_metadataFetcher_GetLocalMetadataMergesWithEnv(t *testing.T) {
	tmp := t.TempDir()
	_, lastCommit := createLocalRepoWithCommits(t, tmp)
	providedName := "test provided name"

	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("INFRACOST_VCS_COMMIT_AUTHOR_NAME", providedName)

	pullID := "1234"
	t.Setenv("INFRACOST_VCS_PULL_REQUEST_ID", pullID)

	test := false
	m := metadataFetcher{
		mu:     &sync.KeyMutex{},
		client: &http.Client{Timeout: time.Second * 5},
		test:   &test,
	}

	actual, err := m.Get(tmp, nil)
	assert.NoError(t, err)

	assert.Equal(t, Metadata{
		Remote: Remote{
			Host: "github.com",
			Name: "git-fixtures/basic",
			URL:  "https://github.com/git-fixtures/basic.git",
		},
		Branch: Branch{
			Name: "master",
		},
		Commit: Commit{
			SHA:         lastCommit.Hash.String(),
			AuthorName:  providedName,
			AuthorEmail: lastCommit.Author.Email,
			Time:        lastCommit.Author.When,
			Message:     lastCommit.Message,
			ChangedObjects: []string{
				filepath.Join(tmp, "added-file"),
			},
		},
		PullRequest: &PullRequest{ID: pullID, VCSProvider: "github"},
		Pipeline:    nil,
	}, actual)
}

func createLocalRepoWithCommits(t *testing.T, tmp string) (*git.Repository, *object.Commit) {
	t.Helper()
	r, err := git.PlainInit(tmp, false)
	require.NoError(t, err)

	_, err = r.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"https://github.com/git-fixtures/basic.git"},
	})
	require.NoError(t, err)

	w, err := r.Worktree()
	require.NoError(t, err)

	createCommit(t, r, w, tmp, "example-git-file", "hello-world!")
	obj := createCommit(t, r, w, tmp, "added-file", "i'm added!")
	return r, obj
}

func Test_metadataFetcher_Get_ReturnsUserDefinedEnvs(t *testing.T) {
	tt := time.Unix(1661350126, 0).UTC()

	t.Setenv("INFRACOST_VCS_PULL_REQUEST_ID", "INFRACOST_VCS_PULL_REQUEST_ID_VALUE")
	t.Setenv("INFRACOST_VCS_PROVIDER", "github")
	t.Setenv("INFRACOST_VCS_PULL_REQUEST_TITLE", "INFRACOST_VCS_PULL_REQUEST_TITLE_VALUE")
	t.Setenv("INFRACOST_VCS_PULL_REQUEST_AUTHOR", "INFRACOST_VCS_PULL_REQUEST_AUTHOR_VALUE")
	t.Setenv("INFRACOST_VCS_PULL_REQUEST_LABELS", "INFRACOST_VCS_PULL_REQUEST_LABELS_VALUE_1, INFRACOST_VCS_PULL_REQUEST_LABELS_VALUE_2")
	t.Setenv("INFRACOST_VCS_BRANCH", "INFRACOST_BRANCH_VALUE")
	t.Setenv("INFRACOST_VCS_BASE_BRANCH", "INFRACOST_VCS_BASE_BRANCH_VALUE")
	t.Setenv("INFRACOST_VCS_PULL_REQUEST_URL", "https://github.com/infracost/test-repo/pull/1979")
	t.Setenv("INFRACOST_VCS_COMMIT_SHA", "INFRACOST_COMMIT_VALUE")
	t.Setenv("INFRACOST_VCS_COMMIT_AUTHOR_NAME", "INFRACOST_COMMIT_AUTHOR_NAME_VALUE")
	t.Setenv("INFRACOST_VCS_COMMIT_AUTHOR_EMAIL", "INFRACOST_COMMIT_AUTHOR_EMAIL_VALUE")
	timestamp := fmt.Sprintf("%d", tt.Unix())
	t.Setenv("INFRACOST_VCS_COMMIT_TIMESTAMP", timestamp)
	t.Setenv("INFRACOST_VCS_COMMIT_MESSAGE", "INFRACOST_COMMIT_MESSAGE_VALUE")
	t.Setenv("INFRACOST_VCS_REPOSITORY_URL", "https://github.com/infracost/test-repo.git")
	t.Setenv("INFRACOST_VCS_BRANCH", "INFRACOST_BRANCH_VALUE")
	t.Setenv("INFRACOST_VCS_PIPELINE_RUN_ID", "INFRACOST_VCS_PIPELINE_RUN_ID_VALUE")

	test := false
	m := metadataFetcher{
		mu:     &sync.KeyMutex{},
		client: &http.Client{Timeout: time.Second * 5},
		test:   &test,
	}

	actual, _ := m.Get(t.TempDir(), nil)

	_, err := json.Marshal(actual)
	assert.NoError(t, err)

	assert.Equal(t, Metadata{
		Remote: Remote{
			Host: "github.com",
			Name: "infracost/test-repo",
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
			Labels:       []string{"INFRACOST_VCS_PULL_REQUEST_LABELS_VALUE_1", "INFRACOST_VCS_PULL_REQUEST_LABELS_VALUE_2"},
			SourceBranch: "INFRACOST_BRANCH_VALUE",
			BaseBranch:   "INFRACOST_VCS_BASE_BRANCH_VALUE",
			URL:          "https://github.com/infracost/test-repo/pull/1979",
		},
		Pipeline: &Pipeline{ID: "INFRACOST_VCS_PIPELINE_RUN_ID_VALUE"},
	}, actual)
}

func Test_metadataFetcher_Get_ReturnsPRIDFromURL(t *testing.T) {
	t.Setenv("INFRACOST_VCS_PULL_REQUEST_URL", "https://github.com/infracost/test-repo/pull/1979")

	test := false
	m := metadataFetcher{
		mu:     &sync.KeyMutex{},
		client: &http.Client{Timeout: time.Second * 5},
		test:   &test,
	}

	actual, _ := m.Get(t.TempDir(), nil)

	_, err := json.Marshal(actual)
	assert.NoError(t, err)

	assert.Equal(t, Metadata{
		PullRequest: &PullRequest{
			ID:  "1979",
			URL: "https://github.com/infracost/test-repo/pull/1979",
		},
	}, actual)
}

func createCommit(t *testing.T, r *git.Repository, w *git.Worktree, tmp, name, contents string) *object.Commit {
	t.Helper()

	filename := filepath.Join(tmp, name)
	err := os.WriteFile(filename, []byte(contents), 0600)
	require.NoError(t, err)

	_, err = w.Add(name)
	require.NoError(t, err)

	commit, err := w.Commit(name, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	obj, err := r.CommitObject(commit)
	require.NoError(t, err)

	return obj
}
