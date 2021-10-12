package usage

import (
	"bytes"

	yamlv3 "gopkg.in/yaml.v3"
)

const yamlCommentMark = "00__"

// markNodeAsComment marks a node as a comment which we then post process later to add the #
// We could use the yamlv3 FootComment/LineComment but this gets complicated with indentation
// especially when we have edge cases like resources that are fully commented out
func markNodeAsComment(node *yamlv3.Node) {
	node.Value = yamlCommentMark + node.Value
}

// replaceCommentMarks replaces the comment marks in the YAML document with #
func replaceCommentMarks(b []byte) []byte {
	return bytes.ReplaceAll(b, []byte(yamlCommentMark), []byte("# "))
}
