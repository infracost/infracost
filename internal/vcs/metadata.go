package vcs

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/logging"
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
	mu     *keyMutex
	client *http.Client
}

func newMetadataFetcher() *metadataFetcher {
	return &metadataFetcher{
		mu:     &keyMutex{},
		client: &http.Client{Timeout: time.Second * 5},
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
		logging.Logger.Debug("detected Infracost is running in test most returning stub metadata for vcs system call")
		return StubMetadata, nil
	}

	_, ok := lookupEnv("GITHUB_ACTIONS")
	if ok {
		logging.Logger.Debug("fetching GitHub action vcs metadata")
		return f.getGithubMetadata(path)
	}

	_, ok = lookupEnv("GITLAB_CI")
	if ok {
		logging.Logger.Debug("fetching Gitlab CI vcs metadata")
		return f.getGitlabMetadata(path)
	}

	v, ok := lookupEnv("BUILD_REPOSITORY_PROVIDER")
	if ok {
		if v == "github" {
			logging.Logger.Debug("fetching Github vcs metadata from Azure DevOps pipeline")
			return f.getAzureReposGithubMetadata(path)
		}

		logging.Logger.Debug("fetching Azure Repos vcs metadata")
		return f.getAzureReposMetadata(path)
	}

	_, ok = lookupEnv("BITBUCKET_COMMIT")
	if ok {
		logging.Logger.Debug("fetching Github vcs metadata from Bitbucket pipeline")
		return f.getBitbucketMetadata(path)
	}

	_, ok = lookupEnv("CIRCLECI")
	if ok {
		logging.Logger.Debug("fetching Github vcs metadata from Circle CI")
		return f.getCircleCIMetadata(path)
	}

	logging.Logger.Debug("could not detect a specific CI system, fetching local Git metadata")
	return f.getLocalGitMetadata(path)
}

func lookupEnv(name string) (string, bool) {
	v, ok := os.LookupEnv(name)
	return strings.TrimSpace(strings.ToLower(v)), ok
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
		logging.Logger.Debug("detected merge commit as branch name is HEAD, fetching author commit")

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
			logging.Logger.Debugf("clone directory '%s' is not empty, opening exising directory", clonePath)
			unlock()

			r, err = git.PlainOpen(clonePath)
			if err != nil {
				return Metadata{}, fmt.Errorf("could not open previously cloned path %w", err)
			}
		} else {
			logging.Logger.Debugf("cloning auhor commit into '%s'", clonePath)
			r, err = git.PlainClone(clonePath, false, &git.CloneOptions{
				URL:           gjson.GetBytes(event, "repository.clone_url").String(),
				ReferenceName: plumbing.NewBranchReferenceName(headRef),
				Auth: &githttp.BasicAuth{
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

var (
	mergeCommitRegxp = regexp.MustCompile(`(?i)^merge\s([\d\w]+)\sinto\s[\d\w]+`)
)

func (f *metadataFetcher) getAzureReposMetadata(path string) (Metadata, error) {
	m, err := f.getLocalGitMetadata(path)
	if err != nil {
		return m, fmt.Errorf("AzureRepos metadata error, could not fetch initial metadata from local git %w", err)
	}

	if m.Branch.Name == "HEAD" {
		err := f.transformAzureDevOpsMergeCommit(path, &m)
		if err != nil {
			logging.Logger.WithError(err).Debug("could not transform Azure DevOps merge commit continuing with provided PR values")
		}
	}

	res := f.getAzureRepoPRInfo()

	m.Pipeline = &Pipeline{ID: os.Getenv("BUILD_BUILDID")}
	pullID := os.Getenv("SYSTEM_PULLREQUEST_PULLREQUESTID")
	m.PullRequest = &PullRequest{
		ID:           pullID,
		VCSProvider:  "azure_devops_tfsgit",
		Title:        res.Title,
		Author:       res.CreatedBy.UniqueName,
		SourceBranch: strings.TrimLeft(os.Getenv("SYSTEM_PULLREQUEST_SOURCEBRANCH"), "refs/heads/"),
		BaseBranch:   strings.TrimLeft(os.Getenv("SYSTEM_PULLREQUEST_TARGETBRANCH"), "refs/heads/"),
		URL:          fmt.Sprintf("%s/pullrequest/%s", os.Getenv("SYSTEM_PULLREQUEST_SOURCEREPOSITORYURI"), pullID),
	}
	return m, nil
}

type azurePullRequestResponse struct {
	Repository struct {
		Id      string `json:"id"`
		Name    string `json:"name"`
		Url     string `json:"url"`
		Project struct {
			Id          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Url         string `json:"url"`
			State       string `json:"state"`
			Revision    int    `json:"revision"`
		} `json:"project"`
		RemoteUrl string `json:"remoteUrl"`
	} `json:"repository"`
	PullRequestId int    `json:"pullRequestId"`
	CodeReviewId  int    `json:"codeReviewId"`
	Status        string `json:"status"`
	CreatedBy     struct {
		Id          string `json:"id"`
		DisplayName string `json:"displayName"`
		UniqueName  string `json:"uniqueName"`
		Url         string `json:"url"`
		ImageUrl    string `json:"imageUrl"`
	} `json:"createdBy"`
	CreationDate  time.Time `json:"creationDate"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	SourceRefName string    `json:"sourceRefName"`
	TargetRefName string    `json:"targetRefName"`
	MergeStatus   string    `json:"mergeStatus"`
	MergeId       string    `json:"mergeId"`
}

// getAzureRepoPRInfo attempts to get the azurePullRequestResponse using Azure DevOps Pipeline variables.
// This method is expected to often fail as Azure DevOps requires users to explicitly pass System.AccessToken as
// an env var on the job step.
func (f *metadataFetcher) getAzureRepoPRInfo() azurePullRequestResponse {
	var out azurePullRequestResponse
	systemAccessToken := strings.TrimSpace(os.Getenv("SYSTEM_ACCESSTOKEN"))
	if systemAccessToken == "" {
		logging.Logger.Debug("skipping fetching pr title and author, the required pipeline variable System.AccessToken was not provided as an env var named SYSTEM_ACCESSTOKEN")
		return out
	}

	apiURL := fmt.Sprintf("%s_apis/git/repositories/%s/pullRequests/%s", os.Getenv("SYSTEM_COLLECTIONURI"), os.Getenv("BUILD_REPOSITORY_ID"), os.Getenv("SYSTEM_PULLREQUEST_PULLREQUESTID"))
	req, _ := http.NewRequest(http.MethodGet, apiURL, nil)
	req.SetBasicAuth("azdo", systemAccessToken)

	res, err := f.client.Do(req)
	if err != nil {
		logging.Logger.WithError(err).Debugf("could not fetch Azure DevOps pull request information using URL '%s'", apiURL)
		return out
	}

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		logging.Logger.WithFields(logrus.Fields{
			"status_code": res.StatusCode,
			"response":    string(b),
		}).Debugf("received non 200 status code from Azure DevOps pull request API call to: '%s'", apiURL)
		return out
	}

	err = json.NewDecoder(res.Body).Decode(&out)
	if err != nil {
		logging.Logger.WithError(err).Debugf("could not decode reponse body from Azure DevOps pull request API call to '%s'", apiURL)
	}

	return out
}

// getAzureReposGithubMetadata returns the git metadata for a repository hosted on GitHub, but using
// Azure DevOps Pipelines as a build agent.
//
// We are unable to fetch pull request title and author for repositories using the GitHub <> Azure DevOps
// setup. The relevant information is not provided in the pipeline, and there is no GitHub access token
// provided to fetch this information from the GitHub API.
func (f *metadataFetcher) getAzureReposGithubMetadata(path string) (Metadata, error) {
	m, err := f.getLocalGitMetadata(path)
	if err != nil {
		return m, fmt.Errorf("AzureRepos metadata error, could not fetch initial metadata from local git %w", err)
	}

	if m.Branch.Name == "HEAD" {
		err := f.transformAzureDevOpsMergeCommit(path, &m)
		if err != nil {
			logging.Logger.WithError(err).Debug("could not transform Azure DevOps merge commit continuing with provided PR values")
		}
	}

	m.Pipeline = &Pipeline{ID: os.Getenv("BUILD_BUILDID")}
	pullNumber := os.Getenv("SYSTEM_PULLREQUEST_PULLREQUESTNUMBER")
	m.PullRequest = &PullRequest{
		ID:           pullNumber,
		VCSProvider:  "azure_devops_github",
		SourceBranch: strings.TrimLeft(os.Getenv("SYSTEM_PULLREQUEST_SOURCEBRANCH"), "refs/heads/"),
		BaseBranch:   strings.TrimLeft(os.Getenv("SYSTEM_PULLREQUEST_TARGETBRANCH"), "refs/heads/"),
		URL:          fmt.Sprintf("%s/pulls/%s", os.Getenv("SYSTEM_PULLREQUEST_SOURCEREPOSITORYURI"), pullNumber),
	}

	return m, nil
}

func (f *metadataFetcher) transformAzureDevOpsMergeCommit(path string, m *Metadata) error {
	m.Branch.Name = strings.TrimLeft(os.Getenv("SYSTEM_PULLREQUEST_SOURCEBRANCH"), "refs/heads/")

	matches := mergeCommitRegxp.FindStringSubmatch(m.Commit.Message)
	if len(matches) <= 1 {
		logging.Logger.Debugf("could not find reference to PR commit in merge commit message '%s' using git log strategy", m.Commit.Message)

		commit, err := f.shiftCommit(path)
		if err != nil {
			return fmt.Errorf("failed to find find PR commit after merge commit using git log strategy %w", err)
		}

		m.Commit = commit
		return nil
	}

	commit, err := f.fetchCommit(path, matches[1])
	if err != nil {
		return fmt.Errorf("failed to find non merge commit for Azure DevOps repo %w", err)
	}

	m.Commit = commit

	return nil
}

func (f *metadataFetcher) fetchCommit(path string, hash string) (Commit, error) {
	r, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return Commit{}, fmt.Errorf("could not open git directory to fetch commit %s %w", hash, err)
	}

	c, err := r.CommitObject(plumbing.NewHash(hash))
	if err != nil {
		return Commit{}, fmt.Errorf("failed to fetch commit %s %w", hash, err)
	}

	return commitToMetadata(c), nil
}

func (f *metadataFetcher) shiftCommit(path string) (Commit, error) {
	r, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return Commit{}, fmt.Errorf("could not open git directory to get git log %w", err)
	}

	itr, err := r.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
		All:   true,
	})
	if err != nil {
		return Commit{}, fmt.Errorf("failed to return an git log iterator %w", err)
	}

	for {
		c, err := itr.Next()
		if err != nil {
			return Commit{}, fmt.Errorf("could not get non merge commit from the git log %w", err)
		}

		if !isMergeCommit(c.Message) {
			return commitToMetadata(c), nil
		}

		logging.Logger.Debugf("ignoring commit with msg '%s'", c.Message)
	}
}

func (f *metadataFetcher) getBitbucketMetadata(path string) (Metadata, error) {
	m, err := f.getLocalGitMetadata(path)
	if err != nil {
		return m, fmt.Errorf("BitBucket metadata error, could not fetch initial metadata from local git %w", err)
	}

	m.Pipeline = &Pipeline{ID: os.Getenv("BITBUCKET_BUILD_NUMBER")}
	m.PullRequest = &PullRequest{
		VCSProvider: "bitbucket",
		ID:          os.Getenv("BITBUCKET_PR_ID"),
		// we're unable to fetch these without calling the Bitbucket API endpoint:
		// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/
		// However, calling the API requires authentication with variables that we don't
		// have access to in the pipeline (e.g. username, password for basic auth).
		Title:        "",
		Author:       "",
		SourceBranch: os.Getenv("BITBUCKET_BRANCH"),
		BaseBranch:   os.Getenv("BITBUCKET_PR_DESTINATION_BRANCH"),
		URL:          fmt.Sprintf("%s/pull-requests/%s", os.Getenv("BITBUCKET_GIT_HTTP_ORIGIN"), os.Getenv("BITBUCKET_PR_ID")),
	}

	return m, nil
}

func (f *metadataFetcher) getCircleCIMetadata(path string) (Metadata, error) {
	m, err := f.getLocalGitMetadata(path)
	if err != nil {
		return m, fmt.Errorf("circle CI metadata error, could not fetch initial metadata from local git %w", err)
	}

	m.Pipeline = &Pipeline{ID: os.Getenv("CIRCLE_WORKFLOW_ID")}
	m.PullRequest = &PullRequest{
		VCSProvider:  "circleci",
		ID:           os.Getenv("CIRCLE_PR_NUMBER"),
		SourceBranch: os.Getenv("CIRCLE_BRANCH"),
		URL:          os.Getenv("CIRCLE_PULL_REQUEST"),

		// we're unable to fetch these without calling the GitHub, Bitbucket or Gitlab API respectively.
		// Calling the API requires authentication with variables that we don't have access to in the pipeline.
		// e.g. a token or username, password for basic auth.
		Title:      "",
		Author:     "",
		BaseBranch: "",
	}

	return m, nil
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

func isMergeCommit(message string) bool {
	return strings.Contains(strings.ToLower(message), "merge")
}
