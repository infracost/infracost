package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envFileBlockedVars = []string{
	"GIT_SSH_COMMAND",
	"git_config_count",
	"TERRAGRUNT_TFPATH",
	"ENV",
	"AWS_CONFIG_FILE",
	"CI_API_V4_URL",
	"INFRACOST_PRICING_API_ENDPOINT",
	"PRICING_API_ENDPOINT",
	"INFRACOST_POLICY_V2_API_ENDPOINT",
	"INFRACOST_DASHBOARD_ENDPOINT",
	"DASHBOARD_ENDPOINT",
	"INFRACOST_JSON_DUMP_PATH",
}

var envFileAllowedVars = []string{
	"INFRACOST_TEST_SAFE",
	"INFRACOST_TERRAGRUNT_FLAGS",
	"INFRACOST_GIT_SSL_CAINFO",
	"INFRACOST_ENV",
	"REGISTRY_PROXY",
	"JSON_DUMP_PATH",
}

func writeEnvFile(t *testing.T, dir string) string {
	t.Helper()

	content := "INFRACOST_TEST_PRESET=fromfile\n"
	for _, k := range append(envFileBlockedVars, envFileAllowedVars...) {
		content += k + "=fromfile\n"
		t.Setenv(k, "")
		require.NoError(t, os.Unsetenv(k))
	}
	t.Setenv("INFRACOST_TEST_PRESET", "fromenv")

	path := filepath.Join(dir, ".env")
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
	return path
}

func TestLoadUntrustedEnvFile(t *testing.T) {
	path := writeEnvFile(t, t.TempDir())

	require.NoError(t, loadUntrustedEnvFile(path))

	assert.Equal(t, "fromenv", os.Getenv("INFRACOST_TEST_PRESET"))
	for _, k := range envFileAllowedVars {
		assert.Equal(t, "fromfile", os.Getenv(k), k)
	}
	for _, k := range envFileBlockedVars {
		_, ok := os.LookupEnv(k)
		assert.False(t, ok, k)
	}
}

func TestLoadDotEnvTrusted(t *testing.T) {
	dir := t.TempDir()
	writeEnvFile(t, dir)
	t.Chdir(dir)
	t.Setenv("INFRACOST_TRUST_ENVFILE", "true")

	require.NoError(t, loadDotEnv())

	assert.Equal(t, "fromenv", os.Getenv("INFRACOST_TEST_PRESET"))
	for _, k := range append(envFileBlockedVars, envFileAllowedVars...) {
		assert.Equal(t, "fromfile", os.Getenv(k), k)
	}
}
