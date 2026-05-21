package output

import (
	"fmt"
	"os"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ToMarkdown(t *testing.T) {

	cases := []struct {
		Name     string
		Fixture  string
		Expected string
	}{
		{"NoPreviousBreakdown", "testdata/ToMarkdown_NoPreviousBreakdown.json", "testdata/ToMarkdown_NoPreviousBreakdown.md"},
	}

	for i, test := range cases {
		test := test

		t.Run(fmt.Sprintf("%02d.%s", i+1, test.Name), func(t *testing.T) {
			r, err := Load(test.Fixture)
			require.NoError(t, err)

			expected, err := os.ReadFile(test.Expected)
			require.NoError(t, err)

			actual, err := ToMarkdown(r, Options{ShowAllProjects: true}, MarkdownOptions{})
			require.NoError(t, err)

			assert.Equal(t, string(expected), string(actual.Msg))
		})
	}
}

func Test_ToMarkdown_TruncatesByCharCount(t *testing.T) {
	r, err := Load("testdata/ToMarkdown_NoPreviousBreakdown.json")
	require.NoError(t, err)

	const limit = 500
	actual, err := ToMarkdown(r, Options{ShowAllProjects: true}, MarkdownOptions{MaxMessageChars: limit})
	require.NoError(t, err)

	runes := utf8.RuneCountInString(string(actual.Msg))
	assert.LessOrEqualf(t, runes, limit, "output has %d runes, limit %d", runes, limit)
}

func Test_ToMarkdown_CharLimitCountsRunesNotBytes(t *testing.T) {
	r, err := Load("testdata/ToMarkdown_NoPreviousBreakdown.json")
	require.NoError(t, err)

	// Inject a multi-byte project name so that bytes > runes.
	for i := range r.Projects {
		r.Projects[i].Name = "プロジェクト-" + r.Projects[i].Name
	}

	const limit = 600
	actual, err := ToMarkdown(r, Options{ShowAllProjects: true}, MarkdownOptions{MaxMessageChars: limit})
	require.NoError(t, err)

	runes := utf8.RuneCountInString(string(actual.Msg))
	assert.LessOrEqualf(t, runes, limit, "output has %d runes, limit %d (bytes=%d)", runes, limit, len(actual.Msg))
}
