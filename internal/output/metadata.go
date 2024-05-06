package output

import (
	"time"

	"github.com/infracost/infracost/internal/config"
)

// Metadata holds common information used to identify the system that Infracost is run within.
type Metadata struct {
	InfracostCommand string `json:"infracostCommand"`

	Branch            string    `json:"vcsBranch"`
	BaseCommitSHA     string    `json:"vcsBaseCommitSha,omitempty"`
	CommitSHA         string    `json:"vcsCommitSha"`
	CommitAuthorName  string    `json:"vcsCommitAuthorName"`
	CommitAuthorEmail string    `json:"vcsCommitAuthorEmail"`
	CommitTimestamp   time.Time `json:"vcsCommitTimestamp"`
	CommitMessage     string    `json:"vcsCommitMessage"`

	VCSRepositoryURL     string   `json:"vcsRepositoryUrl,omitempty"`
	VCSProvider          string   `json:"vcsProvider,omitempty"`
	VCSBaseBranch        string   `json:"vcsBaseBranch,omitempty"`
	VCSPullRequestTitle  string   `json:"vcsPullRequestTitle,omitempty"`
	VCSPullRequestURL    string   `json:"vcsPullRequestUrl,omitempty"`
	VCSPullRequestAuthor string   `json:"vcsPullRequestAuthor,omitempty"`
	VCSPullRequestLabels []string `json:"vcsPullRequestLabels,omitempty"`
	VCSPipelineRunID     string   `json:"vcsPipelineRunId,omitempty"`
	VCSPullRequestID     string   `json:"vcsPullRequestId,omitempty"`

	UsageApiEnabled        bool   `json:"usageApiEnabled,omitempty"`
	UsageFilePath          string `json:"usageFilePath,omitempty"`
	ConfigFilePath         string `json:"configFilePath,omitempty"`
	ConfigFileHasUsageFile bool   `json:"configFileHasUsageFile,omitempty"`
}

// NewMetadata returns a Metadata struct filled with information built from the RunContext.
func NewMetadata(ctx *config.RunContext) Metadata {
	m := Metadata{
		InfracostCommand:  ctx.CMD,
		Branch:            ctx.VCSMetadata.Branch.Name,
		BaseCommitSHA:     ctx.VCSMetadata.BaseCommit.SHA,
		CommitSHA:         ctx.VCSMetadata.Commit.SHA,
		CommitAuthorEmail: ctx.VCSMetadata.Commit.AuthorEmail,
		CommitAuthorName:  ctx.VCSMetadata.Commit.AuthorName,
		CommitTimestamp:   ctx.VCSMetadata.Commit.Time.UTC(),
		CommitMessage:     ctx.VCSMetadata.Commit.Message,
		VCSRepositoryURL:  ctx.VCSRepositoryURL(),
	}

	if ctx.VCSMetadata.PullRequest != nil {
		m.VCSProvider = ctx.VCSMetadata.PullRequest.VCSProvider
		m.VCSPullRequestID = ctx.VCSMetadata.PullRequest.ID
		m.VCSPullRequestURL = ctx.VCSMetadata.PullRequest.URL
		m.VCSBaseBranch = ctx.VCSMetadata.PullRequest.BaseBranch
		m.VCSPullRequestTitle = ctx.VCSMetadata.PullRequest.Title
		m.VCSPullRequestAuthor = ctx.VCSMetadata.PullRequest.Author
		m.VCSPullRequestLabels = ctx.VCSMetadata.PullRequest.Labels
	}

	if ctx.VCSMetadata.Pipeline != nil {
		m.VCSPipelineRunID = ctx.VCSMetadata.Pipeline.ID
	}

	if ctx.Config.UsageAPIEndpoint != "" {
		m.UsageApiEnabled = true
	}
	if ctx.Config.UsageFilePath != "" {
		m.UsageFilePath = ctx.Config.UsageFilePath
	}
	if ctx.Config.ConfigFilePath != "" {
		m.ConfigFilePath = ctx.Config.ConfigFilePath
		for _, p := range ctx.Config.Projects {
			if p.UsageFile != "" {
				m.ConfigFileHasUsageFile = true
				break
			}
		}
	}

	return m
}
