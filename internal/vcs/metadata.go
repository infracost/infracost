package vcs

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/tidwall/gjson"
	"github.com/xanzy/go-gitlab"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/sync"
)

var (
	StubMetadata = Metadata{
		Remote: urlStringToRemote("https://github.com/infracost/infracost"),
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

	allowedProviders = map[string]struct{}{
		"github": {}, "gitlab": {}, "azure_repos": {}, "bitbucket": {},
	}
)

func getEnv(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func getEnvList(key string) []string {
	val := getEnv(key)

	if val == "" {
		return nil
	}

	list := make([]string, 0)

	for v := range strings.SplitSeq(val, ",") {
		v = strings.TrimSpace(v)
		if v != "" {
			list = append(list, v)
		}
	}

	return list
}

func buildEnvProvidedMetadata() Metadata {
	remote := urlStringToRemote(getEnv("INFRACOST_VCS_REPOSITORY_URL"))
	m := Metadata{
		Remote: remote,
		Branch: Branch{
			Name: getEnv("INFRACOST_VCS_BRANCH"),
		},
		Commit:      buildDefaultCommit(),
		PullRequest: buildDefaultPR(),
		Pipeline:    &Pipeline{ID: getEnv("INFRACOST_VCS_PIPELINE_RUN_ID")},
	}

	if m.PullRequest != nil && m.PullRequest.VCSProvider == "" {
		m.PullRequest.VCSProvider = vcsProviderFromHost(remote.Host)
	}

	return m
}

func mergeEnvProvidedMetadata(m Metadata) Metadata {
	envMeta := buildEnvProvidedMetadata()
	if envMeta.Remote.Host != "" {
		m.Remote = envMeta.Remote
	}

	m.Branch = Branch{
		Name: useEnvValueOrParsedString(envMeta.Branch.Name, m.Branch.Name),
	}

	m.Commit = Commit{
		SHA:            useEnvValueOrParsedString(envMeta.Commit.SHA, m.Commit.SHA),
		AuthorName:     useEnvValueOrParsedString(envMeta.Commit.AuthorName, m.Commit.AuthorName),
		AuthorEmail:    useEnvValueOrParsedString(envMeta.Commit.AuthorEmail, m.Commit.AuthorEmail),
		Time:           useEnvValueOrParsedTime(envMeta.Commit.Time, m.Commit.Time),
		Message:        useEnvValueOrParsedString(envMeta.Commit.Message, m.Commit.Message),
		ChangedObjects: m.Commit.ChangedObjects,
		GitDiffTarget:  m.Commit.GitDiffTarget,
	}

	if envMeta.PullRequest != nil {
		if m.PullRequest == nil {
			m.PullRequest = envMeta.PullRequest
		} else {
			m.PullRequest = &PullRequest{
				ID:           useEnvValueOrParsedString(envMeta.PullRequest.ID, m.PullRequest.ID),
				VCSProvider:  useEnvValueOrParsedString(envMeta.PullRequest.VCSProvider, m.PullRequest.VCSProvider),
				Title:        useEnvValueOrParsedString(envMeta.PullRequest.Title, m.PullRequest.Title),
				Author:       useEnvValueOrParsedString(envMeta.PullRequest.Author, m.PullRequest.Author),
				SourceBranch: useEnvValueOrParsedString(envMeta.PullRequest.SourceBranch, m.PullRequest.SourceBranch),
				BaseBranch:   useEnvValueOrParsedString(envMeta.PullRequest.BaseBranch, m.PullRequest.BaseBranch),
				URL:          useEnvValueOrParsedString(envMeta.PullRequest.URL, m.PullRequest.URL),
			}
		}
	}

	if m.PullRequest != nil && m.PullRequest.VCSProvider == "" {
		m.PullRequest.VCSProvider = vcsProviderFromHost(m.Remote.Host)
	}

	if envMeta.Pipeline.ID != "" {
		m.Pipeline = &Pipeline{
			ID: envMeta.Pipeline.ID,
		}
	}

	return m
}

func useEnvValueOrParsedTime(envVal time.Time, parsedVal time.Time) time.Time {
	if !envVal.IsZero() {
		return envVal
	}

	return parsedVal
}

func useEnvValueOrParsedString(envVal string, parsedVal string) string {
	if envVal != "" {
		return envVal
	}

	return parsedVal
}

func buildDefaultPR() *PullRequest {
	if !keysSet(
		"INFRACOST_VCS_PULL_REQUEST_ID",
		"INFRACOST_VCS_PROVIDER",
		"INFRACOST_VCS_PULL_REQUEST_TITLE",
		"INFRACOST_VCS_PULL_REQUEST_AUTHOR",
		"INFRACOST_VCS_PULL_REQUEST_LABELS",
		"INFRACOST_VCS_BRANCH",
		"INFRACOST_VCS_BASE_BRANCH",
		"INFRACOST_VCS_PULL_REQUEST_URL",
	) {
		return nil
	}

	provider := strings.ToLower(getEnv("INFRACOST_VCS_PROVIDER"))
	if _, ok := allowedProviders[provider]; !ok && provider != "" {
		logging.Logger.Warn().Msgf("provided value for INFRACOST_VCS_PROVIDER '%s' is not valid. Setting vcsProvider to an empty string", provider)
		provider = ""
	}

	prURL := getEnv("INFRACOST_VCS_PULL_REQUEST_URL")
	prID := getEnv("INFRACOST_VCS_PULL_REQUEST_ID")
	if prURL != "" && prID == "" {
		prID = getLastURLPart(prURL)
	}

	return &PullRequest{
		ID:           prID,
		VCSProvider:  provider,
		Title:        getEnv("INFRACOST_VCS_PULL_REQUEST_TITLE"),
		Author:       getEnv("INFRACOST_VCS_PULL_REQUEST_AUTHOR"),
		Labels:       getEnvList("INFRACOST_VCS_PULL_REQUEST_LABELS"),
		SourceBranch: getEnv("INFRACOST_VCS_BRANCH"),
		BaseBranch:   getEnv("INFRACOST_VCS_BASE_BRANCH"),
		URL:          prURL,
	}
}

func getLastURLPart(url string) string {
	pieces := strings.Split(url, "/")
	if len(pieces) == 0 {
		return url
	}

	return pieces[len(pieces)-1]
}

func keysSet(keys ...string) bool {
	for _, key := range keys {
		if getEnv(key) != "" {
			return true
		}
	}

	return false
}

func buildDefaultCommit() Commit {
	return Commit{
		SHA:         getEnv("INFRACOST_VCS_COMMIT_SHA"),
		AuthorName:  getEnv("INFRACOST_VCS_COMMIT_AUTHOR_NAME"),
		AuthorEmail: getEnv("INFRACOST_VCS_COMMIT_AUTHOR_EMAIL"),
		Time:        envToTime("INFRACOST_VCS_COMMIT_TIMESTAMP"),
		Message:     getEnv("INFRACOST_VCS_COMMIT_MESSAGE"),
	}
}

func envToTime(key string) time.Time {
	env := getEnv(key)
	if env == "" {
		return time.Time{}
	}

	i, err := strconv.ParseInt(env, 10, 64)
	if err != nil {
		logging.Logger.Debug().Err(err).Msgf("could not parse 'INFRACOST_COMMIT_TIMESTAMP' value '%s' as int64 timestamp", env)
		return time.Time{}
	}

	t := time.Unix(i, 0)
	// if the year is below these values then the date is considered invalid and will throw a
	// json marshall error. In this situation we just return a zero time.
	if y := t.Year(); y < 0 || y >= 10000 {
		return time.Time{}
	}

	return t.UTC()
}

// metadataFetcher is an object designed to find metadata for different systems.
// It is designed to be safe for parallelism. So interactions across branches and for different commits
// will not affect other goroutines.
type metadataFetcher struct {
	mu     *sync.KeyMutex
	client *http.Client
	test   *bool
}

func newMetadataFetcher() *metadataFetcher {
	return &metadataFetcher{
		mu:     &sync.KeyMutex{},
		client: &http.Client{Timeout: time.Second * 5},
	}
}

func (f *metadataFetcher) isTest() bool {
	if f.test != nil {
		return *f.test
	}

	return getEnv("INFRACOST_ENV") == "test" || strings.HasSuffix(os.Args[0], ".test")
}

// Get fetches VCS metadata for the given environment.
// It takes a path argument which should point to the filesystem directory for that should
// be used as the VCS project. This is normally the path to the `.git` directory. If no `.git`
// directory is found in path, Get will traverse parent directories to try and determine VCS metadata.
// Get also supplements base VCS metadata with CI specific data if it can be found.
//
// If a gitDiffTarget argument is provided Get will use the local git fetcher to try and resolve
// file changes to the target. This is equivalent to a `git diff ${gitDiffTarget}`.
//
// When Get encounters an error fetching metadata it will return a default that contains basic Metadata information.
func (f *metadataFetcher) Get(path string, gitDiffTarget *string) (m Metadata, err error) {
	defer func() {
		if f.isTest() {
			return
		}

		// let's now merge any user provide metadata from environment variables.
		// User provided values will always override the parsed metadata.
		m = mergeEnvProvidedMetadata(m)
	}()

	if f.isTest() {
		logging.Logger.Debug().Msg("returning stub metadata as Infracost is running in test mode")
		return StubMetadata, nil
	}

	v, ok := lookupEnv("GITHUB_ACTIONS")
	if ok && v != "" {
		logging.Logger.Debug().Msg("fetching GitHub action VCS metadata")
		return f.getGithubMetadata(path, gitDiffTarget)
	}

	_, ok = lookupEnv("GITLAB_CI")
	if ok {
		logging.Logger.Debug().Msg("fetching Gitlab CI VCS metadata")
		return f.getGitlabMetadata(path, gitDiffTarget)
	}

	v, ok = lookupEnv("BUILD_REPOSITORY_PROVIDER")
	if ok {
		if v == "github" {
			logging.Logger.Debug().Msg("fetching GitHub VCS metadata from Azure DevOps pipeline")
			return f.getAzureReposGithubMetadata(path, gitDiffTarget)
		}

		logging.Logger.Debug().Msg("fetching Azure Repos VCS metadata")
		return f.getAzureReposMetadata(path, gitDiffTarget)
	}

	_, ok = lookupEnv("BITBUCKET_COMMIT")
	if ok {
		logging.Logger.Debug().Msg("fetching GitHub VCS metadata from Bitbucket pipeline")
		return f.getBitbucketMetadata(path, gitDiffTarget)
	}

	_, ok = lookupEnv("CIRCLECI")
	if ok {
		logging.Logger.Debug().Msg("fetching GitHub VCS metadata from Circle CI")
		return f.getCircleCIMetadata(path, gitDiffTarget)
	}

	ok = lookupEnvPrefix("ATLANTIS_")
	if ok {
		logging.Logger.Debug().Msg("fetching Atlantis VCS metadata")
		return f.getAtlantisMetadata(path, gitDiffTarget)
	}

	logging.Logger.Debug().Msg("could not detect a specific CI system, fetching local Git metadata")
	return f.getLocalGitMetadata(path, gitDiffTarget)
}

func lookupEnvPrefix(name string) bool {
	for _, k := range os.Environ() {
		if strings.HasPrefix(k, name) {
			logging.Logger.Debug().Msgf("os env '%s' matched prefix '%s'", k, name)
			return true
		}
	}

	return false
}

func lookupEnv(name string) (string, bool) {
	v, ok := os.LookupEnv(name)
	return strings.TrimSpace(strings.ToLower(v)), ok
}

func (f *metadataFetcher) getGitlabMetadata(path string, gitDiffTarget *string) (Metadata, error) {
	m, err := f.getLocalGitMetadata(path, gitDiffTarget)
	if err != nil {
		return m, fmt.Errorf("GitLab metadata error, could not fetch initial metadata from local git %w", err)
	}

	if m.Branch.Name == "HEAD" {
		m.Branch.Name = getEnv("CI_MERGE_REQUEST_SOURCE_BRANCH_NAME")
		if m.Branch.Name == "" {
			m.Branch.Name = getEnv("CI_COMMIT_BRANCH")
		}
	}

	prUrl := ""
	if getEnv("CI_MERGE_REQUEST_IID") != "" {
		prUrl = getEnv("CI_PROJECT_URL") + "/merge_requests/" + getEnv("CI_MERGE_REQUEST_IID")
	}

	m.Remote = urlStringToRemote(getEnv("CI_PROJECT_URL"))
	m.Pipeline = &Pipeline{ID: getEnv("CI_PIPELINE_ID")}
	m.PullRequest = &PullRequest{
		VCSProvider:  "gitlab",
		ID:           getEnv("CI_MERGE_REQUEST_IID"),
		Title:        getEnv("CI_MERGE_REQUEST_TITLE"),
		Author:       f.getGitlabPullRequestAuthor(),
		SourceBranch: getEnv("CI_MERGE_REQUEST_SOURCE_BRANCH_NAME"),
		BaseBranch:   getEnv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME"),
		URL:          prUrl,
	}

	return m, nil
}

func (f *metadataFetcher) getGithubMetadata(path string, gitDiffTarget *string) (Metadata, error) {
	event, err := os.ReadFile(getEnv("GITHUB_EVENT_PATH"))
	if err != nil {
		return Metadata{}, fmt.Errorf("could not read the GitHub event file %w", err)
	}

	m, err := f.getLocalGitMetadata(path, gitDiffTarget)
	if err != nil {
		return m, fmt.Errorf("GitHub metadata error, could not fetch initial metadata from local git %w", err)
	}

	// if the branch name is HEAD this means that we're using a merge commit and need
	// to fetch the actual commit.
	if m.Branch.Name == "HEAD" {
		logging.Logger.Debug().Msg("detected merge commit as branch name is HEAD, fetching author commit")

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
			logging.Logger.Debug().Msgf("clone directory '%s' is not empty, opening existing directory", clonePath)
			unlock()

			r, err = git.PlainOpen(clonePath)
			if err != nil {
				return Metadata{}, fmt.Errorf("could not open previously cloned path %w", err)
			}
		} else {
			logging.Logger.Debug().Msgf("cloning author commit into '%s'", clonePath)
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

		m.Commit = commitToMetadata(commit, gitDiffTarget)
		m.Branch.Name = gjson.GetBytes(event, "pull_request.head.ref").String()
	}

	remote := gjson.GetBytes(event, "repository.html_url").String()

	m.Remote = urlStringToRemote(remote)
	m.Pipeline = &Pipeline{ID: getEnv("GITHUB_RUN_ID")}

	l := gjson.GetBytes(event, "pull_request.labels").Array()
	labels := make([]string, 0, len(l))
	for _, v := range l {
		labels = append(labels, v.String())
	}

	m.PullRequest = &PullRequest{
		VCSProvider:  "github",
		ID:           gjson.GetBytes(event, "pull_request.number").String(),
		Title:        gjson.GetBytes(event, "pull_request.title").String(),
		Author:       gjson.GetBytes(event, "pull_request.user.login").String(),
		Labels:       labels,
		SourceBranch: gjson.GetBytes(event, "pull_request.head.ref").String(),
		BaseBranch:   gjson.GetBytes(event, "pull_request.base.ref").String(),
		URL:          gjson.GetBytes(event, "pull_request._links.html.href").String(),
	}

	return m, nil
}

func (f *metadataFetcher) getLocalGitMetadata(path string, gitDiffTarget *string) (Metadata, error) {
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
	rem, err := r.Remote("origin")
	if err != nil {
		logging.Logger.Debug().Err(err).Msg("failed to ls remotes")
	}

	if rem != nil {
		urls := rem.Config().URLs
		if len(urls) > 0 {
			remote = urls[0]
		}
	}

	return Metadata{
		Remote: urlStringToRemote(remote),
		Branch: Branch{Name: branch},
		Commit: commitToMetadata(commit, gitDiffTarget, f.getFileChanges(path, r, commit, gitDiffTarget)...),
	}, nil
}

func (f *metadataFetcher) getFileChanges(path string, r *git.Repository, currentCommit *object.Commit, gitDiffTarget *string) []string {
	changedMap := make(map[string]struct{})
	tree, err := currentCommit.Tree()
	if err != nil {
		logging.Logger.Debug().Err(err).Msg("could not get local git file changes, failed to get the tree of current commit")
		return nil
	}
	var nextTree *object.Tree
	if gitDiffTarget != nil {
		b, err := r.ResolveRevision(plumbing.Revision(*gitDiffTarget))
		if err != nil {
			logging.Logger.Debug().Err(err).Msgf("could not get local git file changes, could not resolve revision to branch %s", *gitDiffTarget)
			return nil
		}

		previousCommit, err := r.CommitObject(*b)
		if err != nil {
			logging.Logger.Debug().Err(err).Msg("could not get local git file changes, failed to get a branch commit to compare the current commit against")
			return nil
		}

		nextTree, err = previousCommit.Tree()
		if err != nil {
			logging.Logger.Debug().Err(err).Msg("could not get local git file changes, failed to get the tree of previous commits")
			return nil
		}
	} else {
		commitIter, err := r.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})
		if err != nil {
			logging.Logger.Debug().Err(err).Msgf("could not get local git file changes, could not call git log using currentCommit hash %s", currentCommit.Hash)
			return nil
		}

		// call Next() twice to get the commit after the currentCommit.
		_, _ = commitIter.Next()
		previousCommit, err := commitIter.Next()
		if err != nil {
			logging.Logger.Debug().Err(err).Msg("could not get local git file changes, failed to get the previous commit to compare the current commit against, this is likely because of a shallow clone in CI")
			return nil
		}

		nextTree, err = previousCommit.Tree()
		if err != nil {
			logging.Logger.Debug().Err(err).Msg("could not get local git file changes, failed to get the tree of previous commit")
			return nil
		}
	}

	changes, err := tree.Diff(nextTree)
	if err != nil {
		logging.Logger.Debug().Err(err).Msg("could not get local git file changes, failed to get the diff between current and previous commit")
		return nil
	}

	for _, change := range changes {
		changedMap[change.From.Name] = struct{}{}
		changedMap[change.To.Name] = struct{}{}
	}

	changedFiles := make([]string, 0, len(changedMap))
	for name := range changedMap {
		if name == "" {
			continue
		}

		if !filepath.IsAbs(name) && !inProject(path, name) {
			name = filepath.Join(path, name)
		}

		changedFiles = append(changedFiles, name)
	}

	sort.Strings(changedFiles)
	return changedFiles
}

func (f *metadataFetcher) getAzureReposMetadata(path string, gitDiffTarget *string) (Metadata, error) {
	m, err := f.getLocalGitMetadata(path, gitDiffTarget)
	if err != nil {
		return m, fmt.Errorf("Azure Repos metadata error, could not fetch initial metadata from local git %w", err)
	}

	if strings.ToLower(m.Branch.Name) == "head" {
		err := f.transformAzureDevOpsMergeCommit(path, &m)
		if err != nil {
			logging.Logger.Debug().Err(err).Msg("could not transform Azure DevOps merge commit continuing with provided PR values")
		}
	}

	res := f.getAzureReposPRInfo()

	m.Remote = urlStringToRemote(os.Getenv("BUILD_REPOSITORY_URI"))
	m.Pipeline = &Pipeline{ID: getEnv("BUILD_BUILDID")}
	pullID := getEnv("SYSTEM_PULLREQUEST_PULLREQUESTID")
	prRemote := urlStringToRemote(getEnv("SYSTEM_PULLREQUEST_SOURCEREPOSITORYURI"))
	m.PullRequest = &PullRequest{
		ID:           pullID,
		VCSProvider:  "azure_devops_tfsgit",
		Title:        res.Title,
		Author:       res.CreatedBy.UniqueName,
		SourceBranch: strings.TrimPrefix(getEnv("SYSTEM_PULLREQUEST_SOURCEBRANCH"), "refs/heads/"),
		BaseBranch:   strings.TrimPrefix(getEnv("SYSTEM_PULLREQUEST_TARGETBRANCH"), "refs/heads/"),
		URL:          fmt.Sprintf("%s/pullrequest/%s", prRemote.URL, pullID),
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

// getAzureReposPRInfo attempts to get the azurePullRequestResponse using Azure DevOps Pipeline variables.
// This method is expected to often fail as Azure DevOps requires users to explicitly pass System.AccessToken as
// an env var on the job step.
func (f *metadataFetcher) getAzureReposPRInfo() azurePullRequestResponse {
	var out azurePullRequestResponse
	systemAccessToken := strings.TrimSpace(getEnv("SYSTEM_ACCESSTOKEN"))
	if systemAccessToken == "" {
		logging.Logger.Debug().Msg("skipping fetching pr title and author, the required pipeline variable System.AccessToken was not provided as an env var named SYSTEM_ACCESSTOKEN")
		return out
	}

	apiURL := fmt.Sprintf("%s_apis/git/repositories/%s/pullRequests/%s", getEnv("SYSTEM_COLLECTIONURI"), getEnv("BUILD_REPOSITORY_ID"), getEnv("SYSTEM_PULLREQUEST_PULLREQUESTID"))
	req, _ := http.NewRequest(http.MethodGet, apiURL, nil)
	req.SetBasicAuth("azdo", systemAccessToken)

	res, err := f.client.Do(req)
	if err != nil {
		logging.Logger.Debug().Err(err).Msgf("could not fetch Azure DevOps pull request information using URL '%s'", apiURL)
		return out
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		logging.Logger.Debug().
			Int("status_code", res.StatusCode).
			Str("response", string(b)).
			Msgf("received non 200 status code from Azure DevOps pull request API call to: '%s'", apiURL)
		return out
	}

	err = json.NewDecoder(res.Body).Decode(&out)
	if err != nil {
		logging.Logger.Debug().Err(err).Msgf("could not decode response body from Azure DevOps pull request API call to '%s'", apiURL)
	}

	return out
}

func (f *metadataFetcher) getAzureReposGithubMetadata(path string, gitDiffTarget *string) (Metadata, error) {
	m, err := f.getLocalGitMetadata(path, gitDiffTarget)
	if err != nil {
		return m, fmt.Errorf("Azure Repos metadata error, could not fetch initial metadata from local git %w", err)
	}

	if strings.ToLower(m.Branch.Name) == "head" {
		err := f.transformAzureDevOpsMergeCommit(path, &m)
		if err != nil {
			logging.Logger.Debug().Err(err).Msg("could not transform Azure DevOps merge commit continuing with provided PR values")
		}
	}

	m.Remote = urlStringToRemote(os.Getenv("BUILD_REPOSITORY_URI"))
	m.Pipeline = &Pipeline{ID: getEnv("BUILD_BUILDID")}
	pullNumber := getEnv("SYSTEM_PULLREQUEST_PULLREQUESTNUMBER")
	m.PullRequest = &PullRequest{
		ID:           pullNumber,
		VCSProvider:  "azure_devops_github",
		SourceBranch: strings.TrimPrefix(getEnv("SYSTEM_PULLREQUEST_SOURCEBRANCH"), "refs/heads/"),
		BaseBranch:   strings.TrimPrefix(getEnv("SYSTEM_PULLREQUEST_TARGETBRANCH"), "refs/heads/"),
		URL:          fmt.Sprintf("%s/pulls/%s", getEnv("SYSTEM_PULLREQUEST_SOURCEREPOSITORYURI"), pullNumber),

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
	m.Branch.Name = strings.TrimPrefix(getEnv("SYSTEM_PULLREQUEST_SOURCEBRANCH"), "refs/heads/")
	if m.Branch.Name == "" {
		m.Branch.Name = getEnv("BUILD_SOURCEBRANCHNAME")
	}

	matches := mergeCommitRegxp.FindStringSubmatch(m.Commit.Message)
	if len(matches) <= 1 {
		logging.Logger.Debug().Msgf("could not find reference to PR commit in merge commit message '%s' using git log strategy", m.Commit.Message)

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

	return commitToMetadata(c, nil), nil
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
			return commitToMetadata(c, nil), nil
		}

		logging.Logger.Debug().Msgf("ignoring commit with msg '%s'", c.Message)
	}
}

func (f *metadataFetcher) getBitbucketMetadata(path string, gitDiffTarget *string) (Metadata, error) {
	m, err := f.getLocalGitMetadata(path, gitDiffTarget)
	if err != nil {
		return m, fmt.Errorf("BitBucket metadata error, could not fetch initial metadata from local git %w", err)
	}

	m.Remote = urlStringToRemote(getEnv("BITBUCKET_GIT_HTTP_ORIGIN"))
	m.Pipeline = &Pipeline{ID: getEnv("BITBUCKET_BUILD_NUMBER")}
	m.PullRequest = &PullRequest{
		VCSProvider:  "bitbucket",
		ID:           getEnv("BITBUCKET_PR_ID"),
		SourceBranch: getEnv("BITBUCKET_BRANCH"),
		BaseBranch:   getEnv("BITBUCKET_PR_DESTINATION_BRANCH"),
		URL:          fmt.Sprintf("%s/pull-requests/%s", getEnv("BITBUCKET_GIT_HTTP_ORIGIN"), getEnv("BITBUCKET_PR_ID")),

		// we're unable to fetch these without calling the Bitbucket API endpoint:
		// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/
		// However, calling the API requires authentication with variables that we don't
		// have access to in the pipeline (e.g. username, password for basic auth).
		Title:  "",
		Author: "",
	}

	return m, nil
}

func (f *metadataFetcher) getCircleCIMetadata(path string, gitDiffTarget *string) (Metadata, error) {
	m, err := f.getLocalGitMetadata(path, gitDiffTarget)
	if err != nil {
		return m, fmt.Errorf("circle CI metadata error, could not fetch initial metadata from local git %w", err)
	}

	m.Remote = urlStringToRemote(getEnv("CIRCLE_REPOSITORY_URL"))
	m.Pipeline = &Pipeline{ID: getEnv("CIRCLE_WORKFLOW_ID")}
	m.PullRequest = &PullRequest{
		VCSProvider:  vcsProviderFromHost(m.Remote.Host),
		ID:           getEnv("CIRCLE_PR_NUMBER"),
		SourceBranch: getEnv("CIRCLE_BRANCH"),
		URL:          getEnv("CIRCLE_PULL_REQUEST"),

		// we're unable to fetch these without calling the GitHub, Bitbucket or Gitlab API respectively.
		// Calling the API requires authentication with variables that we don't have access to in the pipeline.
		// e.g. a token or username, password for basic auth.
		Title:      "",
		Author:     "",
		BaseBranch: "",
	}

	return m, nil
}

func (f *metadataFetcher) getAtlantisMetadata(path string, gitDiffTarget *string) (Metadata, error) {
	m, err := f.getLocalGitMetadata(path, gitDiffTarget)
	if err != nil {
		return m, fmt.Errorf("atlantis metadata error, could not fetch initial metadata from local git %w", err)
	}

	// Atlantis doesn't expose any unique identifiers that we can use to distinguish a pipeline run.
	m.Pipeline = &Pipeline{ID: ""}

	m.PullRequest = &PullRequest{
		ID:           getEnv("PULL_NUM"),
		Author:       getEnv("PULL_AUTHOR"),
		SourceBranch: getEnv("HEAD_BRANCH_NAME"),
		BaseBranch:   getEnv("BASE_BRANCH_NAME"),
		// Atlantis doesn't provide any indication of which VCS system triggers a build.
		// So we build the URL using the remote that the local git config points to.
		URL: getAtlantisPullRequestURL(m.Remote),

		// We're unable to fetch title without calling the GitHub, Bitbucket or Gitlab API respectively.
		// This could be possible using the Atlantis Repo tokens, but it has been left out for now.
		Title: "",
	}

	return m, nil
}

func (f *metadataFetcher) getGitlabPullRequestAuthor() string {
	author := getEnv("CI_COMMIT_AUTHOR")

	token := os.Getenv("GITLAB_TOKEN")
	if token == "" {
		return author
	}

	client, err := gitlab.NewClient(os.Getenv("GITLAB_TOKEN"), f.gitlabClientOps()...)
	if err != nil {
		logging.Logger.Debug().Err(err).Msg("failed to init gitlab client, returning commit author")
		return author
	}

	projectID := os.Getenv("CI_PROJECT_ID")
	mergeRequestID := os.Getenv("CI_MERGE_REQUEST_IID")
	id, err := strconv.Atoi(mergeRequestID)
	if err != nil {
		logging.Logger.Debug().Err(err).Msgf("failed to convert gitlab merge request iid %q to int, returning commit author", mergeRequestID)
		return author
	}

	mergeRequest, _, err := client.MergeRequests.GetMergeRequest(projectID, id, nil)
	if err != nil {
		logging.Logger.Debug().Err(err).Msgf("failed to lookup gitlab merge request '%d' for project %q, returning commit author", id, projectID)
		return author
	}

	if mergeRequest.Author == nil {
		logging.Logger.Debug().Msgf("get merge request '%d' for project %q returned nil author, returning commit author", id, projectID)
		return author
	}

	return mergeRequest.Author.Username
}

func (f *metadataFetcher) gitlabClientOps() []gitlab.ClientOptionFunc {
	var ops []gitlab.ClientOptionFunc
	v := os.Getenv("CI_API_V4_URL")
	if v == "" {
		return ops
	}

	u, err := url.Parse(v)
	if err != nil {
		logging.Logger.Debug().Err(err).Msgf("could not parse %q as a valid URL", v)
		return ops
	}

	ops = append(ops, gitlab.WithBaseURL(fmt.Sprintf("%s://%s", u.Scheme, u.Host)))
	return ops
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

func getAtlantisPullRequestURL(remote Remote) string {
	owner := getEnv("BASE_REPO_OWNER")
	project := getEnv("BASE_REPO_NAME")
	pullNumber := getEnv("PULL_NUM")

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

func commitToMetadata(commit *object.Commit, gitDiffTarget *string, changes ...string) Commit {
	return Commit{
		SHA:            commit.Hash.String(),
		AuthorName:     commit.Author.Name,
		AuthorEmail:    commit.Author.Email,
		Time:           commit.Author.When,
		Message:        strings.TrimRight(commit.Message, "\n\r "),
		ChangedObjects: changes,
		GitDiffTarget:  gitDiffTarget,
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

	ChangedObjects []string
	GitDiffTarget  *string
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
	Labels       []string
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
	Name string
	URL  string
}

// Metadata holds a snapshot of information for a given environment and VCS system.
// PullRequest and Pipeline properties are only populated if running in a CI system.
type Metadata struct {
	Remote      Remote
	Branch      Branch
	Commit      Commit
	BaseCommit  Commit
	PullRequest *PullRequest
	Pipeline    *Pipeline
}

func (m Metadata) HasChanges() bool {
	return len(m.Commit.ChangedObjects) > 0
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
			Host: u.Hostname(),
			Name: generateRepoName(u.Host, u.Path),
			URL:  fmt.Sprintf("https://%s%s", u.Host, u.Path),
		}
	}

	logging.Logger.Debug().Err(err).Msgf("parsing remote '%s' as SCP string", remote)

	match := scpSyntax.FindAllStringSubmatch(remote, -1)
	if len(match) == 0 {
		logging.Logger.Debug().Msg("remote did not match SCP regexp")
		return Remote{}
	}

	m := match[0]
	if len(m) < 4 {
		logging.Logger.Debug().Msg("SCP remote was malformed")
		return Remote{}
	}

	host := m[2]
	path := strings.TrimSuffix(m[3], "/")
	port := ""

	// Check if m[3] is port
	if _, err := strconv.ParseInt(path, 10, 64); err == nil {
		port = ":" + path
		path = m[4]
	}

	if strings.Contains(host, "azure") {
		host = strings.TrimPrefix(m[2], "ssh.")
		path = versionRegxp.ReplaceAllString(path, "")
		pieces := strings.Split(path, "/")
		if len(pieces) == 3 {
			path = strings.Join([]string{pieces[0], pieces[1], "_git", pieces[2]}, "/")
		}
	}

	return Remote{
		Host: host,
		Name: generateRepoName(host, path),
		URL:  fmt.Sprintf("https://%s%s/%s", host, port, path),
	}
}

// generateRepoName returns a repo name generated from the remote URL's path.
// Host is used for Azure cloud detection as it requires additional formatting.
func generateRepoName(host, path string) string {
	name := strings.TrimPrefix(path, "/")
	name = strings.TrimSuffix(name, ".git")

	if strings.Contains(strings.ToLower(host), "azure") {
		name = strings.ReplaceAll(name, "/_git", "")
	}

	return name
}

func inProject(dir string, change string) bool {
	rel, err := filepath.Rel(dir, change)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}
