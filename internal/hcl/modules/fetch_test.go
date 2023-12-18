package modules

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransformSSHToHTTPS(t *testing.T) {
	testCases := []struct {
		sshURL   *url.URL
		expected string
	}{
		{&url.URL{Scheme: "ssh", User: url.User("git"), Path: "user/repo.git", Host: "github.com"}, "https://github.com/user/repo.git"},
		{&url.URL{Scheme: "https", Path: "user/repo.git", Host: "github.com"}, "https://github.com/user/repo.git"},
		{&url.URL{Scheme: "git::ssh", User: url.User("git"), Path: "user/repo.git", Host: "github.com"}, "https://github.com/user/repo.git"},
	}

	for _, tc := range testCases {
		transformed, err := TransformSSHToHttps(tc.sshURL)
		assert.NoError(t, err)

		assert.Equal(t, tc.expected, transformed.String())
	}
}
