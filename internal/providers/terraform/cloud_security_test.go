package terraform

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDirProvider_tokenForRemoteHost verifies that the primary Terraform Cloud
// token is only sent to the trusted host, and never to a host derived from the
// (untrusted) remote-run output URL.
func TestDirProvider_tokenForRemoteHost(t *testing.T) {
	t.Run("returns token when host matches the default", func(t *testing.T) {
		p := &DirProvider{TerraformCloudToken: "primary-token"}
		token, err := p.tokenForRemoteHost("app.terraform.io")
		require.NoError(t, err)
		assert.Equal(t, "primary-token", token)
	})

	t.Run("returns token when host matches the configured host", func(t *testing.T) {
		p := &DirProvider{TerraformCloudToken: "primary-token", TerraformCloudHost: "tfe.mycorp.com"}
		token, err := p.tokenForRemoteHost("tfe.mycorp.com")
		require.NoError(t, err)
		assert.Equal(t, "primary-token", token)
	})

	t.Run("errors on non-default host when no host is configured", func(t *testing.T) {
		p := &DirProvider{TerraformCloudToken: "primary-token"}
		_, err := p.tokenForRemoteHost("attacker.example.com")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "attacker.example.com")
		assert.Contains(t, err.Error(), "TERRAFORM_CLOUD_HOST")
	})

	t.Run("errors on host mismatch when a host is configured", func(t *testing.T) {
		p := &DirProvider{TerraformCloudToken: "primary-token", TerraformCloudHost: "tfe.mycorp.com"}
		_, err := p.tokenForRemoteHost("attacker.example.com")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "attacker.example.com")
		assert.Contains(t, err.Error(), "tfe.mycorp.com")
	})
}

// TestIsTrustedRegistryHost verifies the allowlist used to decide whether the
// TG_TF_REGISTRY_TOKEN may be attached to a registry request.
func TestIsTrustedRegistryHost(t *testing.T) {
	assert.True(t, isTrustedRegistryHost("registry.terraform.io"))
	assert.True(t, isTrustedRegistryHost("registry.opentofu.org"))
	assert.True(t, isTrustedRegistryHost("app.terraform.io"))
	assert.True(t, isTrustedRegistryHost("REGISTRY.TERRAFORM.IO"), "host matching should be case-insensitive")

	assert.False(t, isTrustedRegistryHost("attacker.example.com"))
	assert.False(t, isTrustedRegistryHost("registry.terraform.io.attacker.example.com"))

	t.Run("honors TERRAFORM_CLOUD_HOST", func(t *testing.T) {
		t.Setenv("TERRAFORM_CLOUD_HOST", "tfe.mycorp.com")
		assert.True(t, isTrustedRegistryHost("tfe.mycorp.com"))
		assert.False(t, isTrustedRegistryHost("attacker.example.com"))
	})
}
