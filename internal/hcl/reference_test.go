package hcl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

func Test_ReferenceParsing(t *testing.T) {
	cases := []struct {
		input    []string
		expected string
	}{
		{
			input:    []string{"module", "my-mod"},
			expected: "module.my-mod",
		},
		{
			input:    []string{"aws_s3_bucket", "test"},
			expected: "aws_s3_bucket.test",
		},
		{
			input:    []string{"resource", "aws_s3_bucket", "test"},
			expected: "aws_s3_bucket.test",
		},
		{
			input:    []string{"module", "my-mod"},
			expected: "module.my-mod",
		},
		{
			input:    []string{"data", "aws_iam_policy_document", "s3_policy"},
			expected: "data.aws_iam_policy_document.s3_policy",
		},
		{
			input:    []string{"provider", "aws"},
			expected: "provider.aws",
		},
		{
			input:    []string{"output", "something"},
			expected: "output.something",
		},
	}

	for _, test := range cases {
		t.Run(test.expected, func(t *testing.T) {
			ref, err := newReference(test.input)
			require.NoError(t, err)
			assert.Equal(t, test.expected, ref.String())
		})
	}
}

func TestReference_SetKey(t *testing.T) {
	cases := []struct {
		input    cty.Value
		expected string
	}{
		{
			input:    cty.NullVal(cty.String),
			expected: "test.test",
		},
		{
			input:    cty.NilVal,
			expected: "test.test",
		},
		{
			input:    cty.StringVal("0"),
			expected: "test.test[\"0\"]",
		},
	}

	for _, test := range cases {
		t.Run(test.expected, func(t *testing.T) {
			ref, err := newReference([]string{"resource", "test", "test"})
			require.NoError(t, err)

			ref.SetKey(test.input)
			assert.Equal(t, test.expected, ref.String())
		})
	}
}
