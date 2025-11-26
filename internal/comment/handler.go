package comment

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/infracost/infracost/internal/logging"
)

var defaultTag = "infracost-comment"
var validAtTagKey = "valid-at"

// Comment is an interface that represents a comment on any platform. It wraps
// the platform specific comment structures and is used to abstract the
// logic for finding, creating, updating, and deleting the comments.
type Comment interface {
	// Body returns the body of the comment.
	Body() string

	// Ref returns the reference of the comment, this can be a URL to the HTML page of the comment.
	Ref() string

	// Less compares the comment to another comment and returns true if this
	// comment should be sorted before the other comment.
	Less(c Comment) bool

	// IsHidden returns true if the comment is hidden or minimized.
	IsHidden() bool

	// ValidAt returns the time at which the comment is valid.
	// This is used to determine if a comment should be updated or not.
	ValidAt() *time.Time
}

// PlatformHandler is an interface that represents a platform specific handler.
// It is used to call the platform-specific APIs for finding, creating, updating
// and deleting comments.
type PlatformHandler interface {
	// CallFindMatchingComments calls the platform-specific API to find
	// comments that match the given tag, which has been embedded at the beginning
	// of the comment.
	CallFindMatchingComments(ctx context.Context, tag string) ([]Comment, error)

	// CallCreateComment calls the platform-specific API to create a new comment.
	CallCreateComment(ctx context.Context, body string) (Comment, error)

	// CallUpdateComment calls the platform-specific API to update the body of a comment.
	CallUpdateComment(ctx context.Context, comment Comment, body string) error

	// CallDeleteComment calls the platform-specific API to delete the comment.
	CallDeleteComment(ctx context.Context, comment Comment) error

	// CallHideComment calls the platform-specific API to minimize the comment.
	// This functionality is not supported by all platforms, in which case this
	// will throw a NotImplemented error.
	CallHideComment(ctx context.Context, comment Comment) error

	// AddMarkdownTag adds a tag to the given string.
	AddMarkdownTags(s string, tags []CommentTag) (string, error)
}

// PostResult is a struct that contains the result of posting a comment.
type PostResult = struct {
	// Posted is true if the comment was actually posted.
	Posted bool
	// SkipReason is the reason why the comment was not posted.
	SkipReason string
}

// CommentHandler contains the logic for finding, creating, updating and deleting comments
// on any platform. It uses a PlatformHandler to call the platform-specific APIs.
type CommentHandler struct { //nolint
	PlatformHandler PlatformHandler
	Tag             string
}

// NewCommentHandler creates a new CommentHandler.
func NewCommentHandler(ctx context.Context, platformHandler PlatformHandler, tag string) *CommentHandler {
	if tag == "" {
		tag = defaultTag
	}

	return &CommentHandler{
		PlatformHandler: platformHandler,
		Tag:             tag,
	}
}

type CommentOpts struct {
	ValidAt    *time.Time
	SkipNoDiff bool
}

type CommentTag struct {
	Key   string
	Value string
}

// CommentWithBehavior parses the behavior and calls the corresponding *Comment method. Returns
// boolean indicating if the comment was actually posted.
func (h *CommentHandler) CommentWithBehavior(ctx context.Context, behavior, body string, opts *CommentOpts) (PostResult, error) {
	var result PostResult
	var err error

	backoff := time.Second

	const maxRetries = 5
	for i := range maxRetries {
		result, err = h.commentWithBehaviour(ctx, behavior, body, opts)
		if err == nil {
			return result, nil
		}

		if i == maxRetries-1 {
			break
		}

		logging.Logger.Debug().Err(err).Msgf("received an error trying to post comment pausing %v seconds then will retry", backoff.Seconds())

		time.Sleep(backoff)
		backoff *= 2
	}

	return result, err
}
func (h *CommentHandler) commentWithBehaviour(ctx context.Context, behavior string, body string, opts *CommentOpts) (PostResult, error) {
	switch behavior {
	case "update":
		return h.UpdateComment(ctx, body, opts)
	case "new":
		err := h.NewComment(ctx, body, opts)
		if err != nil {
			return PostResult{Posted: false}, err
		}

		return PostResult{Posted: true}, nil
	case "hide-and-new":
		return h.HideAndNewComment(ctx, body, opts)
	case "delete-and-new":
		return h.DeleteAndNewComment(ctx, body, opts)
	}

	return PostResult{Posted: false}, nil
}

// matchingComments returns all comments that match the tag.
func (h *CommentHandler) matchingComments(ctx context.Context) ([]Comment, error) {
	logging.Logger.Info().Msgf("Finding matching comments for tag %s", h.Tag)

	matchingComments, err := h.PlatformHandler.CallFindMatchingComments(ctx, h.Tag)
	if err != nil {
		return nil, h.newPlatformError(err)
	}

	if len(matchingComments) == 1 {
		logging.Logger.Info().Msg("Found 1 matching comment")
	} else {
		logging.Logger.Info().Msgf("Found %d matching comments", len(matchingComments))
	}

	sort.Slice(matchingComments, func(i, j int) bool {
		return matchingComments[i].Less(matchingComments[j])
	})

	return matchingComments, nil
}

// UpdateComment updates the comment with the given body. Returns a PostResult indicating
// if the comment was actually posted and the reason why it was not posted.
func (h *CommentHandler) UpdateComment(ctx context.Context, body string, opts *CommentOpts) (PostResult, error) {
	var validAt *time.Time
	var skipNoDiff bool

	if opts != nil {
		validAt = opts.ValidAt
		skipNoDiff = opts.SkipNoDiff
	}

	tags := []CommentTag{
		{
			Key:   h.Tag,
			Value: "",
		},
	}

	if validAt != nil {
		tags = append(tags, CommentTag{
			Key:   validAtTagKey,
			Value: validAt.Format(time.RFC3339),
		})
	}

	bodyWithTags, err := h.PlatformHandler.AddMarkdownTags(body, tags)
	if err != nil {
		return PostResult{Posted: false}, err
	}

	matchingComments, err := h.matchingComments(ctx)
	if err != nil {
		return PostResult{Posted: false}, err
	}

	if len(matchingComments) > 0 {
		latestMatchingComment := matchingComments[len(matchingComments)-1]

		latestValidAt := latestMatchingComment.ValidAt()
		if validAt != nil && latestValidAt != nil && validAt.Before(*latestValidAt) {
			msg := fmt.Sprintf("Not updating comment since the latest one is newer: %s", latestMatchingComment.Ref())
			logging.Logger.Warn().Msg(msg)
			return PostResult{Posted: false, SkipReason: msg}, nil
		}

		if latestMatchingComment.Body() == bodyWithTags {
			msg := fmt.Sprintf("Not updating comment since the latest one matches exactly: %s", latestMatchingComment.Ref())
			logging.Logger.Info().Msg(msg)
			return PostResult{Posted: false, SkipReason: msg}, nil
		}

		logging.Logger.Info().Msgf("Updating comment %s", latestMatchingComment.Ref())

		err := h.PlatformHandler.CallUpdateComment(ctx, latestMatchingComment, bodyWithTags)
		if err != nil {
			return PostResult{Posted: false}, h.newPlatformError(err)
		}
	} else {
		if skipNoDiff {
			msg := "Not creating initial comment since there is no resource or cost difference"
			logging.Logger.Info().Msg(msg)
			return PostResult{Posted: false, SkipReason: msg}, nil
		}

		logging.Logger.Info().Msg("Creating new comment")

		comment, err := h.PlatformHandler.CallCreateComment(ctx, bodyWithTags)
		if err != nil {
			return PostResult{Posted: false}, h.newPlatformError(err)
		}

		logging.Logger.Info().Msgf("Created new comment %s", comment.Ref())
	}

	return PostResult{Posted: true}, nil
}

// NewComment creates a new comment with the given body.
func (h *CommentHandler) NewComment(ctx context.Context, body string, opts *CommentOpts) error {
	var validAt *time.Time
	if opts != nil {
		validAt = opts.ValidAt

		if opts.SkipNoDiff {
			logging.Logger.Warn().Msg("SkipNoDiff option is not supported for new comments")
		}
	}

	tags := []CommentTag{
		{
			Key:   h.Tag,
			Value: "",
		},
	}

	if validAt != nil {
		tags = append(tags, CommentTag{
			Key:   validAtTagKey,
			Value: validAt.Format(time.RFC3339),
		})
	}

	bodyWithTags, err := h.PlatformHandler.AddMarkdownTags(body, tags)
	if err != nil {
		return err
	}

	logging.Logger.Info().Msg("Creating new comment")

	comment, err := h.PlatformHandler.CallCreateComment(ctx, bodyWithTags)
	if err != nil {
		return h.newPlatformError(err)
	}

	logging.Logger.Info().Msgf("Created new comment: %s", comment.Ref())

	return err
}

// HideAndNewComment hides/minimizes all existing matching comment and creates a new one with the given body. Returns
// a PostResult indicating if the comment was actually posted and the reason why it was not posted.
func (h *CommentHandler) HideAndNewComment(ctx context.Context, body string, opts *CommentOpts) (PostResult, error) {
	var validAt *time.Time
	var skipNoDiff bool

	if opts != nil {
		validAt = opts.ValidAt
		skipNoDiff = opts.SkipNoDiff
	}

	matchingComments, err := h.matchingComments(ctx)
	if err != nil {
		return PostResult{Posted: false}, err
	}

	if len(matchingComments) > 0 && validAt != nil {
		latestMatchingComment := matchingComments[len(matchingComments)-1]

		latestValidAt := latestMatchingComment.ValidAt()
		if latestValidAt != nil && validAt.Before(*latestValidAt) {
			msg := fmt.Sprintf("Not adding a new comment since the latest one is newer: %s", latestMatchingComment.Ref())
			logging.Logger.Warn().Msg(msg)
			return PostResult{Posted: false, SkipReason: msg}, nil
		}
	}

	if len(matchingComments) == 0 && skipNoDiff {
		msg := "Not creating initial comment since there is no resource or cost difference"
		logging.Logger.Info().Msg(msg)
		return PostResult{Posted: false, SkipReason: msg}, nil
	}

	err = h.hideComments(ctx, matchingComments)
	if err != nil {
		return PostResult{Posted: false}, err
	}

	err = h.NewComment(ctx, body, opts)
	if err != nil {
		return PostResult{Posted: false}, err
	}

	return PostResult{Posted: true}, nil
}

// hideComments hides/minimizes all the given comments.
func (h *CommentHandler) hideComments(ctx context.Context, comments []Comment) error {
	visibleComments := []Comment{}

	for _, comment := range comments {
		if !comment.IsHidden() {
			visibleComments = append(visibleComments, comment)
		}
	}

	hiddenCommentCount := len(comments) - len(visibleComments)

	if hiddenCommentCount == 1 {
		logging.Logger.Info().Msg("1 comment is already hidden")
	} else if hiddenCommentCount > 0 {
		logging.Logger.Info().Msgf("%d comments are already hidden", hiddenCommentCount)
	}

	if len(visibleComments) == 1 {
		logging.Logger.Info().Msg("Hiding 1 comment")
	} else {
		logging.Logger.Info().Msgf("Hiding %d comments", len(visibleComments))
	}

	for _, comment := range visibleComments {
		logging.Logger.Info().Msgf("Hiding comment %s", comment.Ref())
		err := h.PlatformHandler.CallHideComment(ctx, comment)
		if err != nil {
			return h.newPlatformError(err)
		}
	}

	return nil
}

// DeleteAndNewComment deletes all existing matching comments and creates a new one with the given body. Returns
// a PostResult indicating if the comment was actually posted and the reason why it was not posted.
func (h *CommentHandler) DeleteAndNewComment(ctx context.Context, body string, opts *CommentOpts) (PostResult, error) {
	var validAt *time.Time
	var skipNoDiff bool

	if opts != nil {
		validAt = opts.ValidAt
		skipNoDiff = opts.SkipNoDiff
	}

	matchingComments, err := h.matchingComments(ctx)
	if err != nil {
		return PostResult{Posted: false}, err
	}

	if len(matchingComments) > 0 && validAt != nil {
		latestMatchingComment := matchingComments[len(matchingComments)-1]

		latestValidAt := latestMatchingComment.ValidAt()
		if latestValidAt != nil && validAt.Before(*latestValidAt) {
			msg := fmt.Sprintf("Not adding a new comment since the latest one is newer: %s", latestMatchingComment.Ref())
			logging.Logger.Warn().Msg(msg)
			return PostResult{Posted: false, SkipReason: msg}, nil
		}
	}

	if len(matchingComments) == 0 && skipNoDiff {
		msg := "Not creating initial comment since there is no resource or cost difference"
		logging.Logger.Info().Msg(msg)
		return PostResult{Posted: false, SkipReason: msg}, nil
	}

	err = h.deleteComments(ctx, matchingComments)
	if err != nil {
		return PostResult{Posted: false}, err
	}

	err = h.NewComment(ctx, body, opts)
	if err != nil {
		return PostResult{Posted: false}, err
	}

	return PostResult{Posted: true}, nil
}

// deleteComments hides/minimizes all the given comments.
func (h *CommentHandler) deleteComments(ctx context.Context, comments []Comment) error {
	if len(comments) == 1 {
		logging.Logger.Info().Msg("Deleting 1 comment")
	} else {
		logging.Logger.Info().Msgf("Deleting %d comments", len(comments))
	}

	for _, comment := range comments {
		logging.Logger.Info().Msgf("Deleting comment %s", comment.Ref())
		err := h.PlatformHandler.CallDeleteComment(ctx, comment)
		if err != nil {
			return h.newPlatformError(err)
		}
	}

	return nil
}

// newPlatformError wraps a platform error with multi-line formatting and a link to the docs
func (h *CommentHandler) newPlatformError(err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s\n%w\n\n%s",
		"The pull request comment was generated successfully but could not be posted:",
		err,
		"See https://infracost.io/docs/troubleshooting/#5-posting-comments for help.")
}
