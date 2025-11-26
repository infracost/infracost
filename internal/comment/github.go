package comment

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/pkg/errors"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// githubComment represents a comment on a GitHub pull request or commit. It
// implements the Comment interface.
type githubComment struct {
	globalID    string
	id          int
	body        string
	createdAt   time.Time
	url         string
	isMinimized bool
}

// Body returns the body of the comment
func (c *githubComment) Body() string {
	return c.body
}

// Ref returns the reference to the comment. For GitHub this is a URL to the
// HTML page of the comment.
func (c *githubComment) Ref() string {
	return c.url
}

// Less compares the comment to another comment and returns true if this
// comment should be sorted before the other comment.
func (c *githubComment) Less(other Comment) bool {
	j := other.(*githubComment)

	if c.createdAt.Format(time.RFC3339) != j.createdAt.Format(time.RFC3339) {
		return c.createdAt.Format(time.RFC3339) < j.createdAt.Format(time.RFC3339)
	}

	if c.globalID != j.globalID {
		return c.globalID < j.globalID
	}

	return c.id < j.id
}

// IsHidden returns true if the comment is hidden or minimized.
func (c *githubComment) IsHidden() bool {
	return c.isMinimized
}

// ValidAt returns the time the comment was tagged as being valid at
func (c *githubComment) ValidAt() *time.Time {
	return extractValidAt(c.Body())
}

// GitHubExtra contains any extra inputs that can be passed to the GitHub comment handlers.
type GitHubExtra struct {
	// APIURL is the URL of the GitHub API. This can be set to a custom URL if
	// using GitHub Enterprise. If not set, the default GitHub API URL will be used.
	APIURL string
	// Token is the GitHub API token.
	Token string
	// Tag used to identify the Infracost comment
	Tag string
	// TLSConfig is the TLS configuration to use when connecting to the GitHub API.
	TLSConfig *tls.Config
}

// splitGitHubProject parses a GitHub project string into its owner and repo parts.
func splitGitHubProject(project string) (string, string, error) {
	parts := strings.SplitN(project, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Invalid GitHub repository name: %s, expecting owner/repo", project)
	}
	return parts[0], parts[1], nil
}

// newGitHubAPIClients creates a v3 GitHub client and a v4 (GraphQL) GitHub client.
// If the apiURL is not set, the default GitHub API URL will be used.
func newGitHubAPIClients(ctx context.Context, token string, apiURL string, tlsConfig *tls.Config) (*github.Client, *githubv4.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	transport := http.DefaultTransport.(*http.Transport)
	transport.TLSClientConfig = tlsConfig
	client := &http.Client{Transport: transport}
	httpCtx := context.WithValue(ctx, oauth2.HTTPClient, client)

	tc := oauth2.NewClient(httpCtx, ts)

	// Handle default GitHub API client
	if apiURL == "" || apiURL == "https://api.github.com" {
		return github.NewClient(tc), githubv4.NewClient(tc), nil
	}

	// Handle GitHub Enterprise API client

	// GitHub Enterprise v3 client needs a base URL and upload URL
	// So we need to parse the API URL and add the necessary parts
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error parsing API URL")
	}

	// Add trailing slash
	if !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}

	// Check if "/api/v3/" already exists in the URL path.
	if strings.HasSuffix(u.Path, "/api/v3/") {
		return nil, nil, fmt.Errorf("the /api/v3/ suffix should not be included in the --github-api-url")
	}

	// Add api to path if it doesn't exist
	if !strings.HasSuffix(u.Path, "/api/") {
		u.Path += "api/"
	}

	apiURL = u.String()

	v3client, err := github.NewEnterpriseClient(apiURL+"v3/", apiURL+"uploads/", tc)
	if err != nil {
		return nil, nil, err
	}

	v4client := githubv4.NewEnterpriseClient(apiURL+"graphql", tc)

	return v3client, v4client, nil
}

// githubPRHandler is a PlatformHandler for GitHub pull requests. It
// implements the PlatformHandler interface and contains the functions
// for finding, creating, updating, deleting comments on GitHub pull requests.
type githubPRHandler struct {
	v4client *githubv4.Client
	v3client *github.Client
	owner    string
	repo     string
	prNumber int
}

// NewGitHubPRHandler creates a new CommentHandler for GitHub pull requests.
func NewGitHubPRHandler(ctx context.Context, project, targetRef string, extra GitHubExtra) (*CommentHandler, error) {
	owner, repo, err := splitGitHubProject(project)
	if err != nil {
		return nil, err
	}

	prNumber, err := strconv.Atoi(targetRef)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing targetRef as pull request number")
	}

	v3client, v4client, err := newGitHubAPIClients(ctx, extra.Token, extra.APIURL, extra.TLSConfig)
	if err != nil {
		return nil, err
	}

	h := &githubPRHandler{
		v3client: v3client,
		v4client: v4client,
		owner:    owner,
		repo:     repo,
		prNumber: prNumber,
	}

	return NewCommentHandler(ctx, h, extra.Tag), nil
}

// CallFindMatchingComments calls the GitHub API to find the pull request
// comments that match the given tag, which has been embedded at the beginning
// of the comment.
func (h *githubPRHandler) CallFindMatchingComments(ctx context.Context, tag string) ([]Comment, error) {
	var q struct {
		Repository struct {
			PullRequest struct {
				Comments struct {
					Nodes []struct {
						ID          githubv4.String
						DatabaseID  int64
						URL         githubv4.String
						CreatedAt   githubv4.DateTime
						PublishedAt githubv4.DateTime
						Body        githubv4.String
						IsMinimized githubv4.Boolean
					}
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"comments(first: 100, after: $after)"`
			} `graphql:"pullRequest(number: $prNumber)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}
	variables := map[string]any{
		"owner":    githubv4.String(h.owner),
		"repo":     githubv4.String(h.repo),
		"prNumber": githubv4.Int(h.prNumber), // nolint:gosec // ignore G115: integer overflow conversion int -> int32
		"after":    (*githubv4.String)(nil),  // Null after argument to get first page.
	}

	// Get comments from all pages.
	var allComments []Comment
	for {
		err := h.v4client.Query(ctx, &q, variables)
		if err != nil {
			return []Comment{}, err
		}
		for _, node := range q.Repository.PullRequest.Comments.Nodes {
			createdAt := node.PublishedAt
			if createdAt.IsZero() {
				createdAt = node.CreatedAt
			}

			allComments = append(allComments, &githubComment{
				globalID:    string(node.ID),
				id:          int(node.DatabaseID),
				body:        string(node.Body),
				createdAt:   createdAt.Time,
				url:         string(node.URL),
				isMinimized: bool(node.IsMinimized),
			})
		}
		if !q.Repository.PullRequest.Comments.PageInfo.HasNextPage {
			break
		}
		variables["after"] = githubv4.NewString(q.Repository.PullRequest.Comments.PageInfo.EndCursor)
	}

	var matchingComments []Comment
	for _, comment := range allComments {
		if hasTagKey(comment.Body(), tag) {
			matchingComments = append(matchingComments, comment)
		}
	}

	return matchingComments, nil
}

// CallCreateComment calls the GitHub API to create a new comment on the pull request.
func (h *githubPRHandler) CallCreateComment(ctx context.Context, body string) (Comment, error) {
	comment, _, err := h.v3client.Issues.CreateComment(
		ctx,
		h.owner,
		h.repo,
		h.prNumber,
		&github.IssueComment{Body: github.String(body)},
	)
	if err != nil {
		return nil, err
	}

	return &githubComment{
		globalID:    comment.GetNodeID(),
		id:          int(comment.GetID()),
		body:        comment.GetBody(),
		createdAt:   comment.GetCreatedAt(),
		url:         comment.GetHTMLURL(),
		isMinimized: false,
	}, nil
}

// CallUpdateComment calls the GitHub API to update the body of a comment on the pull request.
func (h *githubPRHandler) CallUpdateComment(ctx context.Context, comment Comment, body string) error {
	var m struct {
		UpdateIssueComment struct {
			ClientMutationId githubv4.ID //nolint
		} `graphql:"updateIssueComment(input: $input)"`
	}

	input := githubv4.UpdateIssueCommentInput{
		ID:   githubv4.NewString(githubv4.String(comment.(*githubComment).globalID)),
		Body: githubv4.String(body),
	}

	return h.v4client.Mutate(ctx, &m, input, nil)
}

// CallDeleteComment calls the GitHub API to delete the pull request comment.
func (h *githubPRHandler) CallDeleteComment(ctx context.Context, comment Comment) error {
	var m struct {
		DeleteIssueComment struct {
			ClientMutationId githubv4.ID //nolint
		} `graphql:"deleteIssueComment(input: $input)"`
	}

	input := githubv4.DeleteIssueCommentInput{
		ID: githubv4.NewString(githubv4.String(comment.(*githubComment).globalID)),
	}

	return h.v4client.Mutate(ctx, &m, input, nil)
}

// CallHideComment calls the GitHub API to minimize the pull request comment.
func (h *githubPRHandler) CallHideComment(ctx context.Context, comment Comment) error {
	var m struct {
		MinimizeComment struct {
			ClientMutationId githubv4.ID //nolint
		} `graphql:"minimizeComment(input: $input)"`
	}

	input := githubv4.MinimizeCommentInput{
		SubjectID:  githubv4.NewString(githubv4.String(comment.(*githubComment).globalID)),
		Classifier: githubv4.ReportedContentClassifiersOutdated,
	}

	return h.v4client.Mutate(ctx, &m, input, nil)
}

// AddMarkdownTags prepends tags as a markdown comment to the given string.
func (h *githubPRHandler) AddMarkdownTags(s string, tags []CommentTag) (string, error) {
	return addMarkdownTags(s, tags)
}

// githubCommitHandler is a PlatformHandler for GitHub commits. It
// implements the PlatformHandler interface and contains the functions
// for finding, creating, updating, deleting and hiding comments on GitHub commits.
type githubCommitHandler struct {
	v4client  *githubv4.Client
	v3client  *github.Client
	owner     string
	repo      string
	commitSHA string
}

// NewGitHubCommitHandler creates a new PlatformHandler for GitHub commits.
func NewGitHubCommitHandler(ctx context.Context, project, targetRef string, extra GitHubExtra) (*CommentHandler, error) {
	owner, repo, err := splitGitHubProject(project)
	if err != nil {
		return nil, err
	}

	v3client, v4client, err := newGitHubAPIClients(ctx, extra.Token, extra.APIURL, extra.TLSConfig)
	if err != nil {
		return nil, err
	}

	h := &githubCommitHandler{
		v3client:  v3client,
		v4client:  v4client,
		owner:     owner,
		repo:      repo,
		commitSHA: targetRef,
	}

	return NewCommentHandler(ctx, h, extra.Tag), nil
}

// CallFindMatchingComments calls the GitHub API to find the commit
// comments that match the given tag, which has been embedded at the beginning
// of the comment.
func (h *githubCommitHandler) CallFindMatchingComments(ctx context.Context, tag string) ([]Comment, error) {
	var q struct {
		Repository struct {
			Object struct {
				Commit struct {
					Comments struct {
						Nodes []struct {
							ID          githubv4.String
							DatabaseID  int64
							URL         githubv4.String
							CreatedAt   githubv4.DateTime
							PublishedAt githubv4.DateTime
							Body        githubv4.String
							IsMinimized githubv4.Boolean
						}
						PageInfo struct {
							EndCursor   githubv4.String
							HasNextPage bool
						}
					} `graphql:"comments(first: 100, after: $after)"`
				} `graphql:"...on Commit"`
			} `graphql:"object(oid: $commitSha)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]any{
		"owner":     githubv4.String(h.owner),
		"repo":      githubv4.String(h.repo),
		"commitSha": githubv4.GitObjectID(h.commitSHA),
		"after":     (*githubv4.String)(nil), // Null after argument to get first page.
	}

	// Get comments from all pages.
	var allComments []Comment
	for {
		err := h.v4client.Query(ctx, &q, variables)
		if err != nil {
			return []Comment{}, err
		}
		for _, commentNode := range q.Repository.Object.Commit.Comments.Nodes {
			createdAt := commentNode.PublishedAt
			if createdAt.IsZero() {
				createdAt = commentNode.CreatedAt
			}

			allComments = append(allComments, &githubComment{
				globalID:    string(commentNode.ID),
				id:          int(commentNode.DatabaseID),
				body:        string(commentNode.Body),
				createdAt:   createdAt.Time,
				url:         string(commentNode.URL),
				isMinimized: bool(commentNode.IsMinimized),
			})
		}
		if !q.Repository.Object.Commit.Comments.PageInfo.HasNextPage {
			break
		}
		variables["after"] = githubv4.NewString(q.Repository.Object.Commit.Comments.PageInfo.EndCursor)
	}

	var matchingComments []Comment
	for _, comment := range allComments {
		if hasTagKey(comment.Body(), tag) {
			matchingComments = append(matchingComments, comment)
		}
	}

	return matchingComments, nil
}

// CallCreateComment calls the GitHub API to create a new comment on the commit.
func (h *githubCommitHandler) CallCreateComment(ctx context.Context, body string) (Comment, error) {
	comment, _, err := h.v3client.Repositories.CreateComment(
		ctx,
		h.owner,
		h.repo,
		h.commitSHA,
		&github.RepositoryComment{Body: github.String(body)},
	)
	if err != nil {
		return nil, err
	}

	return &githubComment{
		globalID:    comment.GetNodeID(),
		id:          int(comment.GetID()),
		body:        comment.GetBody(),
		createdAt:   comment.GetCreatedAt(),
		url:         comment.GetHTMLURL(),
		isMinimized: false,
	}, nil
}

// CallUpdateComment calls the GitHub API to update the body of a comment on the commit.
func (h *githubCommitHandler) CallUpdateComment(ctx context.Context, comment Comment, body string) error {
	_, _, err := h.v3client.Repositories.UpdateComment(
		ctx,
		h.owner,
		h.repo,
		int64(comment.(*githubComment).id),
		&github.RepositoryComment{Body: github.String(body)},
	)

	if err != nil {
		return err
	}

	return nil
}

// CallDeleteComment calls the GitHub API to delete the commit comment.
func (h *githubCommitHandler) CallDeleteComment(ctx context.Context, comment Comment) error {
	_, err := h.v3client.Repositories.DeleteComment(
		ctx,
		h.owner,
		h.repo,
		int64(comment.(*githubComment).id),
	)

	if err != nil {
		return err
	}

	return nil
}

// CallHideComment calls the GitHub API to minimize the commit comment.
func (h *githubCommitHandler) CallHideComment(ctx context.Context, comment Comment) error {
	var m struct {
		MinimizeComment struct {
			ClientMutationId githubv4.ID //nolint
		} `graphql:"minimizeComment(input: $input)"`
	}

	input := githubv4.MinimizeCommentInput{
		SubjectID:  githubv4.NewString(githubv4.String(comment.(*githubComment).globalID)),
		Classifier: githubv4.ReportedContentClassifiersOutdated,
	}

	return h.v4client.Mutate(ctx, &m, input, nil)
}

// AddMarkdownTag prepends tags as a markdown comment to the given string.
func (h *githubCommitHandler) AddMarkdownTags(s string, tags []CommentTag) (string, error) {
	return addMarkdownTags(s, tags)
}
