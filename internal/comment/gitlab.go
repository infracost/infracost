package comment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
)

// gitlabComment represents a comment on a GitLab merge request or commit. It
// implements the Comment interface.
type gitlabComment struct {
	id           string
	body         string
	createdAt    string
	url          string
	discussionID string
}

// Body returns the body of the comment
func (c *gitlabComment) Body() string {
	return c.body
}

// Ref returns the reference to the comment. For GitLab this is a URL to the
// HTML page of the comment.
func (c *gitlabComment) Ref() string {
	return c.url
}

// Less compares the comment to another comment and returns true if this
// comment should be sorted before the other comment.
func (c *gitlabComment) Less(other Comment) bool {
	j := other.(*gitlabComment)

	if c.createdAt != j.createdAt {
		return c.createdAt < j.createdAt
	}

	return c.id < j.id
}

// IsHidden always returns false for GitLab since GitLab doesn't have a
// feature for hiding comments.
func (c *gitlabComment) IsHidden() bool {
	return false
}

// ValidAt returns the time the comment was tagged as being valid at
func (c *gitlabComment) ValidAt() *time.Time {
	return extractValidAt(c.Body())
}

// GitLabExtra contains any extra inputs that can be passed to the GitLab comment handlers.
type GitLabExtra struct {
	// ServerURL is the URL of the GitLab server. This can be set to a custom URL if
	// using GitLab enterprise. If not set, the default GitLab server URL will be used.
	ServerURL string
	// Token is the GitLab API token.
	Token string
	// Tag used to identify the Infracost comment
	Tag string
}

// newGitLabAPIClients creates a HTTP client and a GraphQL client.
// If the serverURL is not set, the default GitLab server URL will be used.
func newGitLabAPIClients(ctx context.Context, token string, serverURL string) (*http.Client, *graphql.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(ctx, ts)

	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error parsing server URL")
	}

	if !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}

	return httpClient, graphql.NewClient(fmt.Sprintf("%sapi/graphql", u.String()), httpClient), nil
}

// gitlabPRHandler is a PlatformHandler for GitLab merge requests. It
// implements the PlatformHandler interface and contains the functions
// for finding, creating, updating, deleting comments on GitLab merge requests.
type gitlabPRHandler struct {
	httpClient    *http.Client
	graphqlClient *graphql.Client
	serverURL     string
	project       string
	mrNumber      int
}

// NewGitLabPRHandler creates a new PlatformHandler for GitLab merge requests.
func NewGitLabPRHandler(ctx context.Context, project string, targetRef string, extra GitLabExtra) (*CommentHandler, error) {
	mrNumber, err := strconv.Atoi(targetRef)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing targetRef as merge request number")
	}

	serverURL := extra.ServerURL

	// Handle default GitLab API client
	if serverURL == "" {
		serverURL = "https://gitlab.com"
	}

	httpClient, graphqlClient, err := newGitLabAPIClients(ctx, extra.Token, serverURL)
	if err != nil {
		return nil, err
	}

	h := &gitlabPRHandler{
		httpClient:    httpClient,
		graphqlClient: graphqlClient,
		serverURL:     serverURL,
		project:       project,
		mrNumber:      mrNumber,
	}

	return NewCommentHandler(ctx, h, extra.Tag), nil
}

// CallFindMatchingComments calls the GitLab API to find the merge request
// comments that match the given tag, which has been embedded at the beginning
// of the comment.
func (h *gitlabPRHandler) CallFindMatchingComments(ctx context.Context, tag string) ([]Comment, error) {
	var q struct {
		Project struct {
			MergeRequest struct {
				Notes struct {
					Nodes []struct {
						ID        graphql.String
						URL       graphql.String
						CreatedAt graphql.String
						Body      graphql.String
					}
					PageInfo struct {
						EndCursor   graphql.String
						HasNextPage bool
					}
				} `graphql:"notes(first: 100, after: $after)"`
			} `graphql:"mergeRequest(iid: $mrNumber)"`
		} `graphql:"project(fullPath: $project)"`
	}
	variables := map[string]any{
		"project":  graphql.ID(h.project),
		"mrNumber": graphql.String(strconv.Itoa(h.mrNumber)),
		"after":    (*graphql.String)(nil), // Null after argument to get first page.
	}

	// Get comments from all pages.
	var allComments []Comment
	for {
		err := h.graphqlClient.Query(ctx, &q, variables)
		if err != nil {
			return []Comment{}, err
		}
		for _, node := range q.Project.MergeRequest.Notes.Nodes {
			allComments = append(allComments, &gitlabComment{
				id:        string(node.ID),
				body:      string(node.Body),
				createdAt: string(node.CreatedAt),
				url:       string(node.URL),
			})
		}
		if !q.Project.MergeRequest.Notes.PageInfo.HasNextPage {
			break
		}
		variables["after"] = q.Project.MergeRequest.Notes.PageInfo.EndCursor
	}

	var matchingComments []Comment
	for _, comment := range allComments {
		if hasTagKey(comment.Body(), tag) {
			matchingComments = append(matchingComments, comment)
		}
	}

	return matchingComments, nil
}

// CallCreateComment calls the GitLab API to create a new comment on the merge request.
func (h *gitlabPRHandler) CallCreateComment(ctx context.Context, body string) (Comment, error) {
	// Use the REST API here. We'd have to do 2 requests for GraphQL to get the Merge Request ID as well
	reqData, err := json.Marshal(map[string]any{
		"body": body,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error marshaling comment body")
	}

	url := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%d/notes", h.serverURL, url.PathEscape(h.project), h.mrNumber)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqData))
	if err != nil {
		return nil, errors.Wrap(err, "Error creating request")
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := h.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating comment")
	}

	if res.StatusCode != http.StatusCreated {
		return nil, errors.Errorf("Error creating comment: %s", res.Status)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading response body")
	}

	var resData = struct {
		ID        int    `json:"id"`
		CreatedAt string `json:"created_at"`
		Body      string `json:"body"`
	}{}

	err = json.Unmarshal(resBody, &resData)
	if err != nil {
		return nil, errors.Wrap(err, "Error unmarshaling response body")
	}

	refURL := fmt.Sprintf("%s/%s/-/merge_requests/%d#note_%d", h.serverURL, h.project, h.mrNumber, resData.ID)

	return &gitlabComment{
		id:        string(strconv.Itoa(resData.ID)),
		body:      resData.Body,
		createdAt: resData.CreatedAt,
		url:       refURL,
	}, nil
}

// CallUpdateComment calls the GitLab API to update the body of a comment on the merge request.
func (h *gitlabPRHandler) CallUpdateComment(ctx context.Context, comment Comment, body string) error {
	var m struct {
		UpdateNote struct {
			ClientMutationId graphql.ID //nolint
		} `graphql:"updateNote(input: $input)"`
	}

	type UpdateNoteInput struct {
		ID   graphql.ID     `json:"id"`
		Body graphql.String `json:"body"`
	}

	variables := map[string]any{
		"input": UpdateNoteInput{
			ID:   graphql.String(comment.(*gitlabComment).id),
			Body: graphql.String(body),
		},
	}

	return h.graphqlClient.Mutate(ctx, &m, variables)
}

// CallDeleteComment calls the GitLab API to delete the merge request comment.
func (h *gitlabPRHandler) CallDeleteComment(ctx context.Context, comment Comment) error {
	var m struct {
		DestroyNote struct {
			ClientMutationId graphql.ID //nolint
		} `graphql:"destroyNote(input: $input)"`
	}

	type DestroyNoteInput struct {
		ID graphql.ID `json:"id"`
	}

	variables := map[string]any{
		"input": DestroyNoteInput{
			ID: graphql.String(comment.(*gitlabComment).id),
		},
	}

	return h.graphqlClient.Mutate(ctx, &m, variables)
}

// CallHideComment calls the GitLab API to minimize the merge request comment.
func (h *gitlabPRHandler) CallHideComment(ctx context.Context, comment Comment) error {
	return errors.New("Not implemented")
}

// AddMarkdownTags prepends tags as a markdown comment to the given string.
func (h *gitlabPRHandler) AddMarkdownTags(s string, tags []CommentTag) (string, error) {
	return addMarkdownTags(s, tags)
}

// gitlabCommitHandler is a PlatformHandler for GitLab commits. It
// implements the PlatformHandler interface and contains the functions
// for finding, creating, updating, deleting comments on GitLab commits.
type gitlabCommitHandler struct {
	httpClient    *http.Client
	graphqlClient *graphql.Client
	serverURL     string
	project       string
	commitSHA     string
}

// NewGitLabCommitHandler creates a new PlatformHandler for GitLab commits.
func NewGitLabCommitHandler(ctx context.Context, project string, targetRef string, extra GitLabExtra) (*CommentHandler, error) {
	serverURL := extra.ServerURL

	// Handle default GitLab API client
	if serverURL == "" {
		serverURL = "https://gitlab.com"
	}

	httpClient, graphqlClient, err := newGitLabAPIClients(ctx, extra.Token, serverURL)
	if err != nil {
		return nil, err
	}

	h := &gitlabCommitHandler{
		httpClient:    httpClient,
		graphqlClient: graphqlClient,
		serverURL:     serverURL,
		project:       project,
		commitSHA:     targetRef,
	}

	return NewCommentHandler(ctx, h, extra.Tag), nil
}

// CallFindMatchingComments calls the GitLab API to find the commit
// comments that match the given tag, which has been embedded at the beginning
// of the comment.
func (h *gitlabCommitHandler) CallFindMatchingComments(ctx context.Context, tag string) ([]Comment, error) {
	// Get comments from all pages.
	var allComments []Comment

	page := "1"

	for {
		url := fmt.Sprintf(
			"%s/api/v4/projects/%s/repository/commits/%s/discussions?per_page=100&page=%s",
			h.serverURL, url.PathEscape(h.project), h.commitSHA, page,
		)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return []Comment{}, errors.Wrap(err, "Error creating request")
		}
		req.Header.Set("Content-Type", "application/json")

		res, err := h.httpClient.Do(req)
		if err != nil {
			return []Comment{}, errors.Wrap(err, "Error getting comments")
		}

		if res.StatusCode != http.StatusOK {
			return []Comment{}, errors.Errorf("Error getting comments: %s", res.Status)
		}

		if res.Body != nil {
			defer res.Body.Close()
		}

		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return []Comment{}, errors.Wrap(err, "Error reading response body")
		}

		var resData []struct {
			ID             string `json:"id"`
			IndividualNote bool   `json:"individual_note"`
			Notes          []struct {
				ID        int    `json:"id"`
				Body      string `json:"body"`
				CreatedAt string `json:"created_at"`
			} `json:"notes"`
		}

		err = json.Unmarshal(resBody, &resData)
		if err != nil {
			return []Comment{}, errors.Wrap(err, "Error unmarshaling response body")
		}

		for _, discussion := range resData {
			if !discussion.IndividualNote {
				continue
			}

			for _, note := range discussion.Notes {
				refURL := fmt.Sprintf("%s/%s/-/commits/%s#note_%d", h.serverURL, h.project, h.commitSHA, note.ID)

				allComments = append(allComments, &gitlabComment{
					id:           strconv.Itoa(note.ID),
					body:         note.Body,
					createdAt:    note.CreatedAt,
					url:          refURL,
					discussionID: discussion.ID,
				})
			}
		}

		page = res.Header.Get("X-Next-Page")
		if page == "" {
			break
		}
	}

	var matchingComments []Comment
	for _, comment := range allComments {
		if hasTagKey(comment.Body(), tag) {
			matchingComments = append(matchingComments, comment)
		}
	}

	return matchingComments, nil
}

// CallCreateComment calls the GitLab API to create a new comment on the commit.
func (h *gitlabCommitHandler) CallCreateComment(ctx context.Context, body string) (Comment, error) {
	reqData, err := json.Marshal(map[string]any{
		"note": body,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error marshaling comment body")
	}

	url := fmt.Sprintf("%s/api/v4/projects/%s/repository/commits/%s/comments", h.serverURL, url.PathEscape(h.project), h.commitSHA)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqData))
	if err != nil {
		return nil, errors.Wrap(err, "Error creating request")
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := h.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating comment")
	}

	if res.StatusCode != http.StatusCreated {
		return nil, errors.Errorf("Error creating comment: %s", res.Status)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading response body")
	}

	var resData = struct {
		CreatedAt string `json:"created_at"`
		Body      string `json:"body"`
	}{}

	err = json.Unmarshal(resBody, &resData)
	if err != nil {
		return nil, errors.Wrap(err, "Error unmarshaling response body")
	}

	refURL := fmt.Sprintf("%s/%s/-/commits/%s", h.serverURL, h.project, h.commitSHA)

	return &gitlabComment{
		body:      resData.Body,
		createdAt: resData.CreatedAt,
		url:       refURL,
	}, nil
}

// CallUpdateComment calls the GitLab API to update the body of a comment on the commit.
func (h *gitlabCommitHandler) CallUpdateComment(ctx context.Context, comment Comment, body string) error {
	reqData, err := json.Marshal(map[string]any{
		"body": body,
	})
	if err != nil {
		return errors.Wrap(err, "Error marshaling comment body")
	}

	url := fmt.Sprintf(
		"%s/api/v4/projects/%s/repository/commits/%s/discussions/%s/notes/%s",
		h.serverURL, url.PathEscape(h.project), h.commitSHA,
		comment.(*gitlabComment).discussionID, comment.(*gitlabComment).id,
	)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(reqData))
	if err != nil {
		return errors.Wrap(err, "Error creating request")
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := h.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "Error updating comment")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.Errorf("Error updating comment: %s", res.Status)
	}

	return nil
}

// CallDeleteComment calls the GitLab API to delete the commit comment.
func (h *gitlabCommitHandler) CallDeleteComment(ctx context.Context, comment Comment) error {
	url := fmt.Sprintf(
		"%s/api/v4/projects/%s/repository/commits/%s/discussions/%s/notes/%s",
		h.serverURL, url.PathEscape(h.project), h.commitSHA,
		comment.(*gitlabComment).discussionID, comment.(*gitlabComment).id,
	)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return errors.Wrap(err, "Error creating request")
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := h.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "Error deleting comment")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		return errors.Errorf("Error deleting comment: %s", res.Status)
	}

	return nil
}

// CallHideComment calls the GitLab API to minimize the commit comment.
func (h *gitlabCommitHandler) CallHideComment(ctx context.Context, comment Comment) error {
	return errors.New("Not implemented")
}

// AddMarkdownTag prepends tags as a markdown comment to the given string.
func (h *gitlabCommitHandler) AddMarkdownTags(s string, tags []CommentTag) (string, error) {
	return addMarkdownTags(s, tags)
}
