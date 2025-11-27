package comment

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// azureReposComment represents a comment on an Azure Repos pull request. It
// implements the Comment interface.
type azureReposComment struct {
	id            int64
	content       string
	publishedDate string
	href          string
	threadID      int64
}

// Body returns the body of the comment
func (c *azureReposComment) Body() string {
	return c.content
}

// Ref returns the reference to the comment. For Azure Repos this is an API URL
// of the comment.
func (c *azureReposComment) Ref() string {
	return c.href
}

// ValidAt returns the time the comment was tagged as being valid at
func (c *azureReposComment) ValidAt() *time.Time {
	return extractValidAt(c.Body())
}

// Less compares the comment to another comment and returns true if this
// comment should be sorted before the other comment.
func (c *azureReposComment) Less(other Comment) bool {
	j := other.(*azureReposComment)

	if c.publishedDate != j.publishedDate {
		return c.publishedDate < j.publishedDate
	}

	return c.id < j.id
}

// IsHidden always returns false for Azure Repos since Azure Repos doesn't have a
// feature for hiding comments.
func (c *azureReposComment) IsHidden() bool {
	return false
}

// AzureReposExtra contains any extra inputs that can be passed to the Azure Repos
// comment handlers.
type AzureReposExtra struct {
	// Token is the Azure DevOps access token.
	Token string
	// Tag is used to identify the Infracost comment.
	Tag        string
	InitActive bool
	// Status is the desired comment thread status: "active" (unresolved) or "closed" (resolved).
	// If empty, defaults to "closed" unless InitActive is true.
	Status string
}

// azureAPIComment represents API response structure of Azure Repos comment.
type azureAPIComment struct {
	ID            int64  `json:"id"`
	Content       string `json:"content"`
	PublishedDate string `json:"publishedDate"`
	IsDeleted     bool   `json:"isDeleted"`
	Links         struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
}

// azurePATLength helps to determine if token is an Azure DevOps Personal Access Token.
const azurePATLength = 52

// newAzureReposAPIClient creates a HTTP client.
func newAzureReposAPIClient(ctx context.Context, token string) (*http.Client, error) {
	accessToken, tokenType := token, "Bearer"

	if len(token) == azurePATLength {
		accessToken = base64.StdEncoding.EncodeToString(
			[]byte(fmt.Sprintf(":%s", accessToken)),
		)
		tokenType = "Basic"
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: accessToken,
			TokenType:   tokenType,
		},
	)
	httpClient := oauth2.NewClient(ctx, ts)

	return httpClient, nil
}

// buildAzureAPIURL converts repo URL to repo's API URL.
func buildAzureAPIURL(repoURL string) (string, error) {
	apiURL, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("error parsing repo URL %w", err)
	}

	urlParts := strings.Split(apiURL.Path, "_git/")
	if len(urlParts) != 2 {
		return "", fmt.Errorf("Invalid repo URL format %s. Expected https://dev.azure.com/org/project/_git/repo/", repoURL)
	}

	// The URL can contain `org@` username part. If it's present in the API URL,
	// requests may result with 401 status even with the provided token.
	apiURL.User = nil
	apiURL.Path = fmt.Sprintf("%s_apis/git/repositories/%s", urlParts[0], urlParts[1])
	if !strings.HasSuffix(apiURL.Path, "/") {
		apiURL.Path += "/"
	}

	return apiURL.String(), nil
}

// azureReposPRHandler is a PlatformHandler for Azure Repos pull requests. It
// implements the PlatformHandler interface and contains the functions
// for finding, creating, updating, deleting comments on Azure Repos pull requests.
type azureReposPRHandler struct {
	httpClient   *http.Client
	repoAPIURL   string
	prNumber     int
	initAsActive bool
	status       string
}

// NewAzureReposPRHandler creates a new PlatformHandler for Azure Repos pull requests.
func NewAzureReposPRHandler(ctx context.Context, repoURL string, targetRef string, extra AzureReposExtra) (*CommentHandler, error) {
	prNumber, err := strconv.Atoi(targetRef)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing targetRef as pull request number")
	}

	httpClient, err := newAzureReposAPIClient(ctx, extra.Token)
	if err != nil {
		return nil, err
	}

	apiURL, err := buildAzureAPIURL(repoURL)
	if err != nil {
		return nil, err
	}

	h := &azureReposPRHandler{
		httpClient:   httpClient,
		repoAPIURL:   apiURL,
		prNumber:     prNumber,
		initAsActive: extra.InitActive,
		status:       extra.Status,
	}

	return NewCommentHandler(ctx, h, extra.Tag), nil
}

// CallFindMatchingComments calls the Azure Repos API to find the pull request
// comments that match the given tag, which has been embedded at the beginning
// of the comment.
func (h *azureReposPRHandler) CallFindMatchingComments(ctx context.Context, tag string) ([]Comment, error) {
	url := fmt.Sprintf("%spullRequests/%d/threads?api-version=6.0", h.repoAPIURL, h.prNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []Comment{}, errors.Wrap(err, "Error getting comments")
	}

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

	var resData = struct {
		Value []struct {
			ID        int64             `json:"id"`
			IsDeleted bool              `json:"isDeleted"`
			Comments  []azureAPIComment `json:"comments"`
		} `json:"value"`
	}{}

	err = json.Unmarshal(resBody, &resData)
	if err != nil {
		return nil, errors.Wrap(err, "Error unmarshaling response body")
	}

	// This API request creates comments only at the top-level of threads,
	// so we can always just pull the first comment in the thread.
	var topLevelComments []Comment

	for _, thread := range resData.Value {
		if thread.IsDeleted {
			continue
		}

		for _, comment := range thread.Comments {
			if comment.IsDeleted || !hasTagKey(comment.Content, tag) {
				continue
			}

			topLevelComments = append(topLevelComments, &azureReposComment{
				id:            comment.ID,
				content:       comment.Content,
				href:          comment.Links.Self.Href,
				publishedDate: comment.PublishedDate,
				threadID:      thread.ID,
			})

			break
		}
	}

	return topLevelComments, nil
}

// getThreadStatus determines the thread status to use, prioritizing the explicit status flag
// over the legacy initAsActive flag.
func (h *azureReposPRHandler) getThreadStatus() string {
	if h.status != "" {
		return h.status
	}
	if h.initAsActive {
		return "active"
	}
	return "closed"
}

// CallCreateComment calls the Azure Repos API to create a new comment on the pull request.
func (h *azureReposPRHandler) CallCreateComment(ctx context.Context, body string) (Comment, error) {
	status := h.getThreadStatus()
	reqData, err := json.Marshal(map[string]interface{}{
		"comments": []map[string]interface{}{
			{
				"content":         body,
				"parentCommentId": 0,
				"commentType":     1,
			},
		},
		"status": status,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Error marshaling comment body")
	}

	url := fmt.Sprintf("%spullRequests/%d/threads?api-version=6.0", h.repoAPIURL, h.prNumber)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqData))
	if err != nil {
		return nil, errors.Wrap(err, "Error creating request")
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error creating comment: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("Error reading response body: %w", err)
		}

		return nil, fmt.Errorf("Error creating comment: %s\n%s", res.Status, string(resBody))
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading response body")
	}

	var resData = struct {
		ID       int64             `json:"id"`
		Comments []azureAPIComment `json:"comments"`
	}{}

	err = json.Unmarshal(resBody, &resData)
	if err != nil {
		return nil, errors.Wrap(err, "Error unmarshaling response body")
	}

	if len(resData.Comments) == 0 {
		// This error should never happen because we are creating the thread with a comment
		return nil, errors.Wrap(err, "Failed to create new thread: empty comment list")
	}

	firstComment := resData.Comments[0]

	return &azureReposComment{
		id:            firstComment.ID,
		content:       firstComment.Content,
		href:          firstComment.Links.Self.Href,
		publishedDate: firstComment.PublishedDate,
		threadID:      resData.ID,
	}, nil
}

// updateThreadStatus updates the status of a comment thread.
func (h *azureReposPRHandler) updateThreadStatus(ctx context.Context, threadID int64, status string) error {
	reqData, err := json.Marshal(map[string]interface{}{
		"status": status,
	})
	if err != nil {
		return errors.Wrap(err, "Error marshaling thread status update")
	}

	url := fmt.Sprintf("%spullRequests/%d/threads/%d?api-version=6.0", h.repoAPIURL, h.prNumber, threadID)

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(reqData))
	if err != nil {
		return errors.Wrap(err, "Error creating thread status update request")
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := h.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "Error updating thread status")
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != http.StatusOK {
		resBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("Error updating thread status: %s\n%s", res.Status, string(resBody))
	}

	return nil
}

// CallUpdateComment calls the Azure Repos API to update the body of a comment on the pull request.
func (h *azureReposPRHandler) CallUpdateComment(ctx context.Context, comment Comment, body string) error {
	reqData, err := json.Marshal(map[string]interface{}{
		"content":         body,
		"parentCommentId": 0,
		"commentType":     1,
	})
	if err != nil {
		return errors.Wrap(err, "Error marshaling comment body")
	}

	url := fmt.Sprintf("%s?api-version=6.0", comment.Ref())

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(reqData))
	if err != nil {
		return errors.Wrap(err, "Error creating request")
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := h.httpClient.Do(req)

	if res.Body != nil {
		defer res.Body.Close()
	}

	if err != nil {
		return err
	}

	// Update thread status if a status is specified
	if h.status != "" {
		azureComment, ok := comment.(*azureReposComment)
		if ok {
			err = h.updateThreadStatus(ctx, azureComment.threadID, h.getThreadStatus())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// CallDeleteComment calls the Azure Repos API to delete the pull request comment.
func (h *azureReposPRHandler) CallDeleteComment(ctx context.Context, comment Comment) error {
	url := fmt.Sprintf("%s?api-version=6.0", comment.Ref())

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return errors.Wrap(err, "Error creating request")
	}

	res, err := h.httpClient.Do(req)

	if res.Body != nil {
		defer res.Body.Close()
	}

	return err
}

// CallHideComment calls the Azure Repos API to minimize the pull request comment.
func (h *azureReposPRHandler) CallHideComment(ctx context.Context, comment Comment) error {
	return errors.New("Not implemented")
}

// AddMarkdownTags prepends tags as a markdown comment to the given string.
func (h *azureReposPRHandler) AddMarkdownTags(s string, tags []CommentTag) (string, error) {
	return addMarkdownTags(s, tags)
}
