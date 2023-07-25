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

// addMarkdownTag prepends tags as a markdown comment to the given string.
func addMarkdownTags(s string, tags []CommentTag) (string, error) {
	combined, err := combineTags(tags)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s\n%s", markdownTag(combined), s), nil
}

func combineTags(tags []CommentTag) (string, error) {
	vals := make([]string, len(tags))

	for i, tag := range tags {
		if tag.Key == "" {
			return "", fmt.Errorf("key cannot be blank")
		}

		if strings.Contains(tag.Key, "=") {
			return "", fmt.Errorf("key cannot contain an =")
		}

		if strings.Contains(tag.Key, ",") {
			return "", fmt.Errorf("value cannot contain a comma")
		}

		if strings.Contains(tag.Value, ",") {
			return "", fmt.Errorf("value cannot contain a comma")
		}

		vals[i] = tag.Key

		if tag.Value != "" {
			vals[i] += fmt.Sprintf("=%s", tag.Value)
		}
	}

	return strings.Join(vals, ", "), nil
}

// hasTagKey returns true if the given body contains a tag with the given key.
func hasTagKey(body, key string) bool {
	pattern := fmt.Sprintf(`\[\/\/\]: <> \([^)]*\s*%s[=,\)]+`, key)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}

	match := re.FindString(body)
	return match != ""
}

// extractTagValue extracts the value of a tag from the given body.
// It returns an empty string if the tag is not found.
func extractTagValue(body, key string) string {
	pattern := fmt.Sprintf(`\[\/\/\]: <> \([^)]*\s*%s=([^,\)]+)`, key)
	re, err := regexp.Compile(pattern)
	if err != nil {
		return ""
	}

	matches := re.FindStringSubmatch(body)
	if len(matches) != 2 {
		return ""
	}

	return matches[1]
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
