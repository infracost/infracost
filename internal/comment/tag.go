package comment

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// markdownTag wraps a tag in a markdown comment.
func markdownTag(s string) string {
	return fmt.Sprintf("[//]: <> (%s)", s)
}

// addMarkdownTag prepends a tag as a markdown comment to the given string.
func addMarkdownTag(s, key, value string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("key cannot be blank")
	}

	if strings.Contains(key, "=") {
		return "", fmt.Errorf("key cannot contain an =")
	}

	tag := key
	if value != "" {
		tag = fmt.Sprintf("%s=%s", key, value)
	}

	return fmt.Sprintf("%s\n%s", markdownTag(tag), s), nil
}

// extractValidAt extract the validAt time from valid-at tag
// which found in the comment body
func extractValidAt(body string) *time.Time {
	validAt := extractTagValue(body, validAtTagKey)
	if validAt == "" {
		return nil
	}

	t, err := time.Parse(time.RFC3339, validAt)
	if err != nil {
		return nil
	}

	return &t
}

// extractTagValue extracts the value of a tag from the given body.
// It returns an empty string if the tag is not found.
func extractTagValue(body, key string) string {
	r, err := regexp.Compile(fmt.Sprintf(`\[\/\/\]: <> \(%s=(.+)\)`, key))
	if err != nil {
		return ""
	}

	matches := r.FindStringSubmatch(body)
	if len(matches) != 2 {
		return ""
	}

	return matches[1]
}
