package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNameFromRepoURL(t *testing.T) {
	tests := []struct {
		repoURL string
		name    string
	}{
		{"git@github.com:org/repo.git", "org/repo"},
		{"https://github.com/org/repo.git", "org/repo"},
		{"git@gitlab.com:org/repo.git", "org/repo"},
		{"https://gitlab.com/org/repo.git", "org/repo"},
		{"git@bitbucket.org:org/repo.git", "org/repo"},
		{"https://user@bitbucket.org/org/repo.git", "org/repo"},
		{"https://user@dev.azure.com/org/project/_git/repo", "org/project/repo"},
		{"git@ssh.dev.azure.com:v3/org/project/repo", "org/project/repo"},
	}

	for _, test := range tests {
		actual := nameFromRepoURL(test.repoURL)
		assert.Equal(t, test.name, actual)
	}
}
