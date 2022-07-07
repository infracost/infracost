package vcs

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/tidwall/gjson"
)

var (
	StubMetadata = Metadata{
		Branch: Branch{
			Name: "stub-branch",
		},
		Commit: Commit{
			SHA:         "stub-sha",
			AuthorName:  "stub-author",
			AuthorEmail: "stub@stub.com",
			Time:        time.Time{},
			Message:     "stub-message",
		},
	}

	MetadataFetcher = newMetadataFetcher()
)

type keyMutex struct {
	mutexes sync.Map // Zero value is empty and ready for use
}

func (m *keyMutex) Lock(key string) func() {
	value, _ := m.mutexes.LoadOrStore(key, &sync.Mutex{})
	mtx := value.(*sync.Mutex)
	mtx.Lock()

	return func() { mtx.Unlock() }
}

// metadataFetcher is an object designed to find metadata for different systems.
// It is designed to be safe for parallelism. So interactions across branches and for different commits
// will not affect other goroutines.
type metadataFetcher struct {
	mu *keyMutex
}

func newMetadataFetcher() *metadataFetcher {
	return &metadataFetcher{
		mu: &keyMutex{},
	}
}

// Get fetches vcs metadata for the given environment.
// It takes a path argument which should point to the filesystem directory for that should
// be used as the VCS project. This is normally the path to the `.git` directory. If no `.git`
// directory is found in path, Get will traverse parent directories to try and determine VCS metadata.
//
// Get also supplements base VCS metadata with CI specific data if it can be found.
func (f *metadataFetcher) Get(path string) (Metadata, error) {
	if isTest() {
		return StubMetadata, nil
	}

	_, ok := os.LookupEnv("GITHUB_ACTIONS")
	if ok {
		return f.getGithubMetadata(path)
	}

	_, ok = os.LookupEnv("GITLAB_CI")
	if ok {
		return f.getGitlabMetadata(path)
	}

	return f.getLocalGitMetadata(path)
}

func isTest() bool {
	return os.Getenv("INFRACOST_ENV") == "test" || strings.HasSuffix(os.Args[0], ".test")
}

func (f *metadataFetcher) getGitlabMetadata(path string) (Metadata, error) {
	m, err := f.getLocalGitMetadata(path)
	if err != nil {
		return m, fmt.Errorf("GitLab metadata error, could not fetch initial metadata from local git %w", err)
	}

	if m.Branch.Name == "HEAD" {
		m.Branch.Name = os.Getenv("CI_MERGE_REQUEST_SOURCE_BRANCH_NAME")
	}

	m.Pipeline = &Pipeline{ID: os.Getenv("CI_PIPELINE_ID")}
	m.PullRequest = &PullRequest{
		VCSProvider:  "gitlab",
		ID:           os.Getenv("CI_MERGE_REQUEST_IID"),
		Title:        os.Getenv("CI_MERGE_REQUEST_TITLE"),
		Author:       os.Getenv("CI_COMMIT_AUTHOR"),
		SourceBranch: os.Getenv("CI_MERGE_REQUEST_SOURCE_BRANCH_NAME"),
		BaseBranch:   os.Getenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME"),
		URL:          os.Getenv("CI_MERGE_REQUEST_PROJECT_URL") + "/-/merge_requests/" + os.Getenv("CI_MERGE_REQUEST_IID"),
	}

	return m, nil
}

func (f *metadataFetcher) getGithubMetadata(path string) (Metadata, error) {
	event, err := os.ReadFile(os.Getenv("GITHUB_EVENT_PATH"))
	if err != nil {
		return Metadata{}, fmt.Errorf("could not read the GitHub event file %w", err)
	}

	m, err := f.getLocalGitMetadata(path)
	if err != nil {
		return m, fmt.Errorf("GitHub metadata error, could not fetch initial metadata from local git %w", err)
	}

	// if the branch name is HEAD this means that we're using a merge commit and need
	// to fetch the actual commit.
	if m.Branch.Name == "HEAD" {
		r, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
		if err != nil {
			return Metadata{}, fmt.Errorf("could not open git directory to fetch metadata %w", err)
		}

		cnf, err := r.Config()
		if err != nil {
			return Metadata{}, fmt.Errorf("error opening the .git/config folder in github %w", err)
		}

		auth := cnf.Raw.Section("http").Subsection("https://github.com/").Options.Get("extraheader")
		val := strings.TrimSpace(strings.ReplaceAll(auth, "AUTHORIZATION: basic", ""))
		b, err := base64.URLEncoding.DecodeString(val)
		if err != nil {
			return Metadata{}, fmt.Errorf("GitHub basic auth credentials were malformed, could not decode %w", err)
		}

		pieces := strings.Split(string(b), ":")
		if len(pieces) != 2 {
			return Metadata{}, fmt.Errorf("GitHub basic auth credentials were malformed, invalid auth components %+v", pieces)
		}

		headRef := gjson.GetBytes(event, "pull_request.head.ref").String()
		clonePath := fmt.Sprintf("/tmp/infracost-%s-%s", gjson.GetBytes(event, "repository.name").String(), headRef)
		unlock := f.mu.Lock(clonePath)

		// if the clone path already exists then let's just do a plain open. We might hit this
		// if the user is running multiple Infracost commands on the head commit. We don't want
		// to clone each time.
		_, err = os.Stat(clonePath)
		if err == nil {
			unlock()

			r, err = git.PlainOpen(clonePath)
			if err != nil {
				return Metadata{}, fmt.Errorf("could not open previously cloned path %w", err)
			}
		} else {
			r, err = git.PlainClone(clonePath, false, &git.CloneOptions{
				URL:           gjson.GetBytes(event, "repository.clone_url").String(),
				ReferenceName: plumbing.NewBranchReferenceName(headRef),
				Auth: &http.BasicAuth{
					Username: pieces[0],
					Password: pieces[1],
				},
				SingleBranch: true,
				Depth:        1,
			})
			unlock()

			if err != nil {
				return Metadata{}, fmt.Errorf("could not shallow clone GitHub repo to fetch commit information %w", err)
			}
		}

		head, err := r.Head()
		if err != nil {
			return Metadata{}, fmt.Errorf("could not determine head from cloned GitHub branch %w", err)
		}

		commit, err := r.CommitObject(head.Hash())
		if err != nil {
			return Metadata{}, fmt.Errorf("could not read head commit from cloned GitHub repo %w", err)
		}

		m.Commit = commitToMetadata(commit)
		m.Branch.Name = gjson.GetBytes(event, "pull_request.head.ref").String()
	}

	m.Pipeline = &Pipeline{ID: os.Getenv("GITHUB_RUN_ID")}
	m.PullRequest = &PullRequest{
		VCSProvider:  "github",
		ID:           gjson.GetBytes(event, "pull_request.number").String(),
		Title:        gjson.GetBytes(event, "pull_request.title").String(),
		Author:       gjson.GetBytes(event, "pull_request.user.login").String(),
		SourceBranch: gjson.GetBytes(event, "pull_request.head.ref").String(),
		BaseBranch:   gjson.GetBytes(event, "pull_request.base.ref").String(),
		URL:          gjson.GetBytes(event, "pull_request._links.html.href").String(),
	}

	return m, nil
}

func (f *metadataFetcher) getLocalGitMetadata(path string) (Metadata, error) {
	r, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return Metadata{}, fmt.Errorf("could not open git directory to fetch metadata %w", err)
	}

	head, err := r.Head()
	if err != nil {
		return Metadata{}, fmt.Errorf("could not determine head from local git directory %w", err)
	}

	branch := head.Name().Short()
	commit, err := r.CommitObject(head.Hash())
	if err != nil {
		return Metadata{}, fmt.Errorf("could not read head commit %w", err)
	}

	return Metadata{
		Branch: Branch{Name: branch},
		Commit: commitToMetadata(commit),
	}, nil
}

func commitToMetadata(commit *object.Commit) Commit {
	return Commit{
		SHA:         commit.Hash.String(),
		AuthorName:  commit.Author.Name,
		AuthorEmail: commit.Author.Email,
		Time:        commit.Author.When,
		Message:     strings.TrimRight(commit.Message, "\n\r "),
	}
}

// Commit defines information for a given commit. This information is normally populated from the
// local vcs environment. Attributes can be overwritten by CI specific properties and variables.
type Commit struct {
	SHA         string
	AuthorName  string
	AuthorEmail string
	Time        time.Time
	Message     string
}

type Branch struct {
	Name string
}

// PullRequest defines information that is unique to a pull request or merge request in a CI system.
type PullRequest struct {
	ID           string
	VCSProvider  string
	Title        string
	Author       string
	SourceBranch string
	BaseBranch   string
	URL          string
}

// Pipeline holds information about a specific run for a CI system.
// This is used to aggregate Infracost metadata across commands used in the same pipeline.
type Pipeline struct {
	ID string
}

// Metadata holds a snapshot of information for a given environment and vcs system.
// PullRequest and Pipeline properties are only populated if running in a CI system.
type Metadata struct {
	Branch      Branch
	Commit      Commit
	PullRequest *PullRequest
	Pipeline    *Pipeline
}
