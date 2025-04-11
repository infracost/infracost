package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactUrl(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "github with secrets",
			url:      "https://user:password@github.com/some/repo",
			expected: "https://****:****@github.com/some/repo",
		},
		{
			name:     "github without secrets",
			url:      "https://github.com/some/repo",
			expected: "https://github.com/some/repo",
		},
		{
			name:     "gitlab ssh with secrets",
			url:      "ssh://user:password@gitlab.com:22",
			expected: "ssh://****:****@gitlab.com:22",
		},
		{
			name:     "gitlab ssh without secrets",
			url:      "ssh://gitlab.com:22",
			expected: "ssh://gitlab.com:22",
		},
		{
			name:     "git error string",
			url:      "Error downloading ssh://user:password@gitlab.com:22 because it couldn't be found",
			expected: "Error downloading ssh://****:****@gitlab.com:22 because it couldn't be found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, RedactUrl(tt.url))
		})
	}
}

func TestRedactUrlPtr(t *testing.T) {
	tests := []struct {
		name     string
		url      *string
		expected *string
	}{
		{
			name:     "nil",
			url:      nil,
			expected: nil,
		},
		{
			name:     "github with secrets",
			url:      stringPtr("https://user:password@github.com/some/repo"),
			expected: stringPtr("https://****:****@github.com/some/repo"),
		},
		{
			name:     "github without secrets",
			url:      stringPtr("https://github.com/some/repo"),
			expected: stringPtr("https://github.com/some/repo"),
		},
		{
			name:     "gitlab ssh with secrets",
			url:      stringPtr("ssh://user:password@gitlab.com:22"),
			expected: stringPtr("ssh://****:****@gitlab.com:22"),
		},
		{
			name:     "gitlab ssh without secrets",
			url:      stringPtr("ssh://gitlab.com:22"),
			expected: stringPtr("ssh://gitlab.com:22"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, RedactUrlPtr(tt.url))
		})
	}

}

func stringPtr(in string) *string {
	return &in
}
