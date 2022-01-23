package comment

import "fmt"

// markdownTag wraps a tag in a markdown comment.
func markdownTag(s string) string {
	return fmt.Sprintf("[//]: <> (%s)", s)
}

// addMarkdownTag prepends a tag as a markdown comment to the given string.
func addMarkdownTag(s string, tag string) string {
	comment := s

	if tag != "" {
		comment = fmt.Sprintf("%s\n%s", markdownTag(tag), comment)
	}

	return comment
}
