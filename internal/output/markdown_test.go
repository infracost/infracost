package output

import (
	"fmt"
	"os"
	"testing"

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
