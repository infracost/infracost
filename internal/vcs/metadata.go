package vcs

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

	scpSyntax        = regexp.MustCompile(`^([a-zA-Z0-9-._~]+@)?([a-zA-Z0-9._-]+):([a-zA-Z0-9./._-]+)(?:\?||$)(.*)$`)
	mergeCommitRegxp = regexp.MustCompile(`(?i)^merge\s([\d\w]+)\sinto\s[\d\w]+`)
	startsWithMerge  = regexp.MustCompile(`(?i)^merge`)
	versionRegxp     = regexp.MustCompile(`^v\d/`)
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

// Get fetches VCS metadata for the given environment.
// It takes a path argument which should point to the filesystem directory for that should
// be used as the VCS project. This is normally the path to the `.git` directory. If no `.git`
// directory is found in path, Get will traverse parent directories to try and determine VCS metadata.
//
// Get also supplements base VCS metadata with CI specific data if it can be found.
func (f *metadataFetcher) Get(path string) (Metadata, error) {
	if isTest() {
		logging.Logger.Debug("returning stub metadata as Infracost is running in test mode")
		return StubMetadata, nil
	}

	_, ok := lookupEnv("GITHUB_ACTIONS")
	if ok {
		logging.Logger.Debug("fetching GitHub action VCS metadata")
		return f.getGithubMetadata(path)
	}

	_, ok = lookupEnv("GITLAB_CI")
	if ok {
		logging.Logger.Debug("fetching Gitlab CI VCS metadata")
		return f.getGitlabMetadata(path)
	}

	v, ok := lookupEnv("BUILD_REPOSITORY_PROVIDER")
	if ok {
		if v == "github" {
			logging.Logger.Debug("fetching GitHub VCS metadata from Azure DevOps pipeline")
			return f.getAzureReposGithubMetadata(path)
		}

		logging.Logger.Debug("fetching Azure Repos VCS metadata")
		return f.getAzureReposMetadata(path)
	}

	_, ok = lookupEnv("BITBUCKET_COMMIT")
	if ok {
		logging.Logger.Debug("fetching GitHub VCS metadata from Bitbucket pipeline")
		return f.getBitbucketMetadata(path)
	}

	_, ok = lookupEnv("CIRCLECI")
	if ok {
		logging.Logger.Debug("fetching GitHub VCS metadata from Circle CI")
		return f.getCircleCIMetadata(path)
	}

	ok = lookupEnvPrefix("ATLANTIS_")
	if ok {
		logging.Logger.Debug("fetching Atlantis VCS metadata")
		return f.getAtlantisMetadata(path)
	}

	_, ok = lookupEnv("TFC_RUN_ID")
	if ok {
		logging.Logger.Debug("fetching TFC run task metadata")
		return f.getTFCMetadata(path)
	}

	logging.Logger.Debug("could not detect a specific CI system, fetching local Git metadata")
	return f.getLocalGitMetadata(path)
}

func lookupEnvPrefix(name string) bool {
	for _, k := range os.Environ() {
		if strings.HasPrefix(k, name) {
			logging.Logger.Debugf("os env '%s' matched prefix '%s'", k, name)
			return true
		}
	}

	return false
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
			logging.Logger.Debugf("clone directory '%s' is not empty, opening existing directory", clonePath)
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

	var remote string
	rms, err := r.Remotes()
	if err != nil {
		logging.Logger.WithError(err).Debug("failed to ls remotes")
	}

	for _, rem := range rms {
		urls := rem.Config().URLs
		if len(urls) > 0 {
			remote = urls[0]
			break
		}
	}

	return Metadata{
		Remote: urlStringToRemote(remote),
		Branch: Branch{Name: branch},
		Commit: commitToMetadata(commit),
	}, nil
}

func (f *metadataFetcher) getAzureReposMetadata(path string) (Metadata, error) {
	m, err := f.getLocalGitMetadata(path)
	if err != nil {
		return m, fmt.Errorf("Azure Repos metadata error, could not fetch initial metadata from local git %w", err)
	}

	if strings.ToLower(m.Branch.Name) == "head" {
		err := f.transformAzureDevOpsMergeCommit(path, &m)
		if err != nil {
			logging.Logger.WithError(err).Debug("could not transform Azure DevOps merge commit continuing with provided PR values")
		}
	}

	res := f.getAzureReposPRInfo()

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
func (f *metadataFetcher) getAzureReposPRInfo() azurePullRequestResponse {
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
		logging.Logger.WithError(err).Debugf("could not decode response body from Azure DevOps pull request API call to '%s'", apiURL)
	}

	return out
}

func (f *metadataFetcher) getAzureReposGithubMetadata(path string) (Metadata, error) {
	m, err := f.getLocalGitMetadata(path)
	if err != nil {
		return m, fmt.Errorf("Azure Repos metadata error, could not fetch initial metadata from local git %w", err)
	}

	if strings.ToLower(m.Branch.Name) == "head" {
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

		// We are unable to fetch pull request title and author for repositories using the GitHub <> Azure DevOps
		// setup. The relevant information is not provided in the pipeline, and there is no GitHub access token
		// provided to fetch this information from the GitHub API.
		Title:  "",
		Author: "",
	}

	return m, nil
}

// transformAzureDevOpsMergeCommit sets the first non merge Commit that metadataFetcher can find on
// Metadata m. transformAzureDevOpsMergeCommit fetches a specific commit sha if it is referenced
// in the original Commit of Metadata m. Otherwise, transformAzureDevOpsMergeCommit returns the first commit
// on a git log call that doesn't appear to be a Merge commit.
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
		VCSProvider:  "bitbucket",
		ID:           os.Getenv("BITBUCKET_PR_ID"),
		SourceBranch: os.Getenv("BITBUCKET_BRANCH"),
		BaseBranch:   os.Getenv("BITBUCKET_PR_DESTINATION_BRANCH"),
		URL:          fmt.Sprintf("%s/pull-requests/%s", os.Getenv("BITBUCKET_GIT_HTTP_ORIGIN"), os.Getenv("BITBUCKET_PR_ID")),

		// we're unable to fetch these without calling the Bitbucket API endpoint:
		// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/
		// However, calling the API requires authentication with variables that we don't
		// have access to in the pipeline (e.g. username, password for basic auth).
		Title:  "",
		Author: "",
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

func (f *metadataFetcher) getAtlantisMetadata(path string) (Metadata, error) {
	m, err := f.getLocalGitMetadata(path)
	if err != nil {
		return m, fmt.Errorf("atlantis metadata error, could not fetch initial metadata from local git %w", err)
	}

	// Atlantis doesn't expose any unique identifiers that we can use to distinguish a pipeline run.
	m.Pipeline = &Pipeline{ID: ""}

	m.PullRequest = &PullRequest{
		VCSProvider:  "atlantis",
		ID:           os.Getenv("PULL_NUM"),
		Author:       os.Getenv("PULL_AUTHOR"),
		SourceBranch: os.Getenv("HEAD_BRANCH_NAME"),
		BaseBranch:   os.Getenv("BASE_BRANCH_NAME"),
		// Atlantis doesn't provide any indication of which VCS system triggers a build.
		// So we build the URL using the remote that the local git config points to.
		URL: getAtlantisPullRequestURL(m.Remote),

		// We're unable to fetch title without calling the GitHub, Bitbucket or Gitlab API respectively.
		// This could be possible using the Atlantis Repo tokens, but it has been left out for now.
		Title: "",
	}

	return m, nil
}

// getTFCMetadata returns Metadata built from a TFC request: https://www.terraform.io/cloud-docs/api-docs/run-tasks/run-tasks-integration#request-body
// We cannot rely on any local git information we run this process in the Infracost Cloud. All OS variables are populated
// by the Infracost Cloud run task worker, which passes them as environment flags when running the Infracost CLI.
func (f *metadataFetcher) getTFCMetadata(path string) (Metadata, error) {
	remote := urlStringToRemote(os.Getenv("INFRACOST_VCS_REPOSITORY_URL"))

	runCreatedAt := os.Getenv("INFRACOST_VCS_COMMIT_CREATED_AT")
	parsedCreatedAt, err := time.Parse(time.RFC3339, runCreatedAt)
	if err != nil {
		logging.Logger.WithError(err).Debugf("could not parse TFC run created time '%s'", runCreatedAt)
	}

	// pullURL is only populated if the run task is triggered by a VCS webhook event. If the run task has been
	// triggered by a manual build in Terraform Cloud then TFC_PULL_URL will be blank. This includes builds that
	// have been originally triggered by a VCS webhook event and then rerun by a user.
	pullURL := os.Getenv("INFRACOST_VCS_PULL_REQUEST_URL")
	return Metadata{
		Remote: remote,
		Branch: Branch{
			Name: os.Getenv("INFRACOST_VCS_SOURCE_BRANCH"),
		},
		Commit: Commit{
			SHA:  getLastURLPart(os.Getenv("INFRACOST_VCS_COMMIT_URL")),
			Time: parsedCreatedAt,
			// we use the TFC_RUN_MESSAGE as the commit message even though this is the PR title.
			// This is consistent with what TFC show, i.e. a commit hash and then a PR title.
			Message: os.Getenv("INFRACOST_VCS_PULL_REQUEST_TITLE"),

			// TFC does not provide us an information on the original VCS author. The only referenced
			// users are TFC users if the run has been triggered by the UI. We leave these fields blank in
			// order to avoid confusion.
			AuthorName:  "",
			AuthorEmail: "",
		},
		PullRequest: &PullRequest{
			ID:           getLastURLPart(pullURL),
			VCSProvider:  vcsProviderFromHost(remote.Host),
			SourceBranch: os.Getenv("INFRACOST_VCS_SOURCE_BRANCH"),
			URL:          pullURL,
			Title:        os.Getenv("INFRACOST_VCS_PULL_REQUEST_TITLE"),

			// TFC does not provide us information on the following fields in the event data that's sent
			// through from the run task.
			Author:     "",
			BaseBranch: "",
		},
		Pipeline: &Pipeline{ID: os.Getenv("TFC_RUN_ID")},
	}, nil
}

func vcsProviderFromHost(host string) string {
	pieces := strings.Split(host, ".")
	if len(pieces) == 2 {
		return pieces[0]
	}

	if len(pieces) > 2 {
		return pieces[len(pieces)-2]
	}

	return host
}

func getLastURLPart(urlString string) string {
	pieces := strings.Split(urlString, "/")

	if len(pieces) == 0 {
		return urlString
	}

	return pieces[len(pieces)-1]
}

func getAtlantisPullRequestURL(remote Remote) string {
	owner := os.Getenv("BASE_REPO_OWNER")
	project := os.Getenv("BASE_REPO_NAME")
	pullNumber := os.Getenv("PULL_NUM")

	if strings.Contains(remote.Host, "github") {
		return fmt.Sprintf("https://%s/%s/%s/pull/%s", remote.Host, owner, project, pullNumber)
	}

	if strings.Contains(remote.Host, "gitlab") {
		return fmt.Sprintf("https://%s/%s/%s/-/merge_requests/%s", remote.Host, owner, project, pullNumber)
	}

	if strings.Contains(remote.Host, "azure") {
		return fmt.Sprintf("https://%s/%s/base/_git/%s/pullrequest/%s", remote.Host, owner, project, pullNumber)
	}

	if strings.Contains(remote.Host, "bitbucket") {
		return fmt.Sprintf("https://%s/%s/%s/pull-requests/%s", remote.Host, owner, project, pullNumber)
	}

	return ""
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
// local VCS environment. Attributes can be overwritten by CI specific properties and variables.
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

// Remote holds information about the upstream repository that the git project uses.
type Remote struct {
	Host string
	URL  string
}

// Metadata holds a snapshot of information for a given environment and VCS system.
// PullRequest and Pipeline properties are only populated if running in a CI system.
type Metadata struct {
	Remote      Remote
	Branch      Branch
	Commit      Commit
	PullRequest *PullRequest
	Pipeline    *Pipeline
}

func isMergeCommit(message string) bool {
	return startsWithMerge.MatchString(message)
}

// urlStringToRemote returns the provided string as a Remote struct, with the host extracted from the URL.
// urlStringToRemote is designed to work with both HTTPS and SCP remote url strings.
func urlStringToRemote(remote string) Remote {
	if remote == "" {
		return Remote{}
	}

	u, err := url.Parse(remote)
	if err == nil {
		return Remote{
			Host: u.Host,
			URL:  fmt.Sprintf("https://%s%s", u.Host, u.Path),
		}
	}

	logging.Logger.WithError(err).Debugf("parsing remote '%s' as SCP string", remote)

	match := scpSyntax.FindAllStringSubmatch(remote, -1)
	if len(match) == 0 {
		logging.Logger.Debug("remote did not match SCP regexp")
		return Remote{}
	}

	m := match[0]
	if len(m) < 4 {
		logging.Logger.Debug("SCP remote was malformed")
		return Remote{}
	}

	host := m[2]
	path := m[3]

	if strings.Contains(host, "azure") {
		host = strings.TrimLeft(m[2], "ssh.")
		path = versionRegxp.ReplaceAllString(path, "")
		pieces := strings.Split(path, "/")
		if len(pieces) == 3 {
			path = strings.Join([]string{pieces[0], pieces[1], "_git", pieces[2]}, "/")
		}
	}

	return Remote{
		Host: host,
		URL:  fmt.Sprintf("https://%s/%s", host, path),
	}
}
