package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadUntrustedEnvFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	content := "INFRACOST_TEST_SAFE=yes\n" +
		"INFRACOST_TEST_PRESET=fromfile\n" +
		"GIT_SSH_COMMAND=touch /tmp/pwned\n" +
		"git_config_count=1\n" +
		"INFRACOST_PRICING_API_ENDPOINT=http://evil\n" +
		"TERRAGRUNT_TFPATH=/tmp/evil\n" +
		"PRICING_API_ENDPOINT=http://evil\n" +
		"INFRACOST_POLICY_V2_API_ENDPOINT=http://evil\n" +
		"AWS_CONFIG_FILE=./evil\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))

	for _, k := range []string{"GIT_SSH_COMMAND", "git_config_count", "INFRACOST_PRICING_API_ENDPOINT", "TERRAGRUNT_TFPATH", "PRICING_API_ENDPOINT", "INFRACOST_POLICY_V2_API_ENDPOINT", "AWS_CONFIG_FILE"} {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}
	t.Setenv("INFRACOST_TEST_SAFE", "")
	os.Unsetenv("INFRACOST_TEST_SAFE")
	t.Setenv("INFRACOST_TEST_PRESET", "fromenv")

	require.NoError(t, loadUntrustedEnvFile(path))

	assert.Equal(t, "yes", os.Getenv("INFRACOST_TEST_SAFE"))
	assert.Equal(t, "fromenv", os.Getenv("INFRACOST_TEST_PRESET"))
	for _, k := range []string{"GIT_SSH_COMMAND", "git_config_count", "INFRACOST_PRICING_API_ENDPOINT", "TERRAGRUNT_TFPATH", "PRICING_API_ENDPOINT", "INFRACOST_POLICY_V2_API_ENDPOINT", "AWS_CONFIG_FILE"} {
		_, ok := os.LookupEnv(k)
		assert.False(t, ok, k)
	}
}
