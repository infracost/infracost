package modules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindLatestMatchingVersion(t *testing.T) {
	versions := []string{
		"1.0.0",
		"1.1.2",
		"1.1.0",
		"1.0.1",
		"1.0.2",
		"1.1.1",
	}

	tests := []struct {
		versions     []string
		constraints  string
		expected     string
		returnsError bool
	}{
		{versions, "", "1.1.2", false},
		{versions, "1.0.2", "1.0.2", false},
		{versions, ">= 1.1.1", "1.1.2", false},
		{versions, ">= 1.0.1, < 1.0.3", "1.0.2", false},
		{versions, "~> 1.0.0", "1.0.2", false},
		{versions, "~> 1.0.0, != 1.0.2", "1.0.1", false},
		{versions, "> 1.1.2", "", true},
		{[]string{}, "1.0.2", "", true},
	}

	for _, test := range tests {
		actual, err := findLatestMatchingVersion(test.versions, test.constraints)
		if test.returnsError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}

		assert.Equal(t, test.expected, actual)
	}
}

func TestNormalizeHost(t *testing.T) {
	tests := []struct {
		host         string
		expected     string
		returnsError bool
	}{
		{"registry.example.com", "registry.example.com", false},
		{"registry.example.com:443", "registry.example.com", false},
		{"registry.example.com:10443", "registry.example.com:10443", false},
		{"https://registry.example.com", "registry.example.com", false},
		{"https://registry.example.com:443", "registry.example.com", false},
		{"https://registry.example.com:10443", "registry.example.com:10443", false},
		{"?invalid?", "", true},
	}

	for _, test := range tests {
		actual, err := normalizeHost(test.host)
		if test.returnsError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}

		assert.Equal(t, test.expected, actual)
	}
}
