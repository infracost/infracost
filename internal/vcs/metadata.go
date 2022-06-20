package vcs

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/tidwall/gjson"
)

var (
	// envToProviderMap specifies unique env vars that determine the CI system being used and maps
	// them to the corresponding function to find vcs metadata.
	envToProviderMap = map[string]GetMetadataFunc{
		"GITHUB_ACTIONS": getGithubMetadata,
		"GITLAB_CI":      getGitlabMetadata,
	}

	StubMetadata = Metadata{
		Branch: Branch{
			Name: "stub-branch",
		},
		Commit: Commit{
			SHA:         "stub-sha",
			AuthorName:  "stub-author",
			AuthorEmail: "stub@stub.com",
			Timestamp:   12345,
			Message:     "stub-message",
		},
	}
)

// GetMetadataFunc defines a function that fetches data for a given vcs implementation.
type GetMetadataFunc func(path string) (Metadata, error)

// GetMetadata fetches vcs metadata for the given environment. GetMetadata will attempt to try and find
// a GetMetadataFunc for the environment using env variables to determine the CI system. If GetMetadata
// cannot determine the CI system it falls back to getting the metadata from the local git filesystem.
func GetMetadata(path string) (Metadata, error) {
	if isTest() {
		return StubMetadata, nil
	}

	for e, f := range envToProviderMap {
		_, ok := os.LookupEnv(e)
		if ok {
			return f(path)
		}
	}

	return getLocalGitMetadata(path)
}

func isTest() bool {
	return os.Getenv("INFRACOST_ENV") == "test" || strings.HasSuffix(os.Args[0], ".test")
}

func getGitlabMetadata(path string) (Metadata, error) {
	m, err := getLocalGitMetadata(path)
	if err != nil {
		return m, fmt.Errorf("GitLab metadata error, could not fetch initial metadata from local git %w", err)
	}

	if m.Branch.Name == "HEAD" {
		m.Branch.Name = os.Getenv("CI_MERGE_REQUEST_SOURCE_BRANCH_NAME")
	}

	m.Pipeline = &Pipeline{ID: os.Getenv("CI_JOB_ID")}
	m.PullRequest = &PullRequest{
		VCSProvider:  "gitlab",
		Title:        os.Getenv("CI_MERGE_REQUEST_TITLE"),
		Author:       os.Getenv("CI_COMMIT_AUTHOR"),
		SourceBranch: os.Getenv("CI_MERGE_REQUEST_SOURCE_BRANCH_NAME"),
		BaseBranch:   os.Getenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME"),
		URL:          os.Getenv("CI_MERGE_REQUEST_PROJECT_URL") + "/-/merge_requests/" + os.Getenv("CI_MERGE_REQUEST_IID"),
	}

	return m, nil
}

func getGithubMetadata(path string) (Metadata, error) {
	event, err := os.ReadFile(os.Getenv("GITHUB_EVENT_PATH"))
	if err != nil {
		return Metadata{}, fmt.Errorf("could not read the GitHub event file %w", err)
	}

	m, err := getLocalGitMetadata(path)
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
		clonePath := fmt.Sprintf("/tmp/%s", headRef)

		// if the clone path already exists then let's just do a plain open. We might hit this
		// if the user is running multiple Infracost commands on the head commit. We don't want
		// to clone each time.
		_, err = os.Stat(clonePath)
		if err == nil {
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
		Title:        gjson.GetBytes(event, "pull_request.title").String(),
		Author:       gjson.GetBytes(event, "pull_request.user.login").String(),
		SourceBranch: gjson.GetBytes(event, "pull_request.head.ref").String(),
		BaseBranch:   gjson.GetBytes(event, "pull_request.base.ref").String(),
		URL:          gjson.GetBytes(event, "pull_request._links.html.href").String(),
	}

	return m, nil
}

func getLocalGitMetadata(path string) (Metadata, error) {
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
		Timestamp:   commit.Author.When.Unix(),
		Message:     commit.Message,
	}
}

// Commit defines information for a given commit. This information is normally populated from the
// local vcs environment. Attributes can be overwritten by CI specific properties and variables.
type Commit struct {
	SHA         string
	AuthorName  string
	AuthorEmail string
	Timestamp   int64
	Message     string
}

type Branch struct {
	Name string
}

// PullRequest defines information that is unique to a pull request or merge request in a CI system.
type PullRequest struct {
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
