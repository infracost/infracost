package hcl

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/extclient"
	"github.com/infracost/infracost/internal/hcl/modules"
	"github.com/infracost/infracost/internal/sync"
)

// blocksFromHCL parses the given Terraform source into a set of Blocks so the
// remote variables loader can be exercised against realistic input.
func blocksFromHCL(t *testing.T, contents string) Blocks {
	t.Helper()

	path := createTestFile("main.tf", contents)
	logger := newDiscardLogger()
	loader := modules.NewModuleLoader(modules.ModuleLoaderOptions{
		CachePath:         filepath.Dir(path),
		HCLParser:         modules.NewSharedHCLParser(),
		CredentialsSource: nil,
		SourceMap:         config.TerraformSourceMap{},
		SourceMapRegex:    nil,
		Logger:            logger,
		ModuleSync:        &sync.KeyMutex{},
	})
	parser := NewParser(
		RootPath{DetectedPath: filepath.Dir(path)},
		CreateEnvFileMatcher([]string{}, nil),
		loader,
		logger,
	)

	module, err := parser.ParseDirectory()
	require.NoError(t, err)

	return module.Blocks
}

// TestTFCRemoteVariablesLoader_Load_HostValidation verifies that the loader
// never sends the configured Terraform Cloud token to a host derived from the
// (untrusted) scanned Terraform when that host does not match the trusted one.
func TestTFCRemoteVariablesLoader_Load_HostValidation(t *testing.T) {
	// A cloud block that points at an attacker-controlled host.
	blocks := blocksFromHCL(t, `
terraform {
  cloud {
    organization = "my-org"
    hostname     = "attacker.example.com"
    workspaces {
      name = "my-workspace"
    }
  }
}
`)

	t.Run("errors when scanned Terraform sets a host and no host is configured", func(t *testing.T) {
		client := extclient.NewAuthedAPIClient("app.terraform.io", "secret-token")
		loader := NewTFCRemoteVariablesLoader(client, "", false, newDiscardLogger())

		_, err := loader.Load(RemoteVarLoaderOptions{Blocks: blocks})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "attacker.example.com")
		assert.Contains(t, err.Error(), "TERRAFORM_CLOUD_HOST")
		// The client must not have been repointed at the untrusted host.
		assert.Equal(t, "app.terraform.io", client.Host())
	})

	t.Run("skips without error when a different trusted host is configured", func(t *testing.T) {
		client := extclient.NewAuthedAPIClient("tfe.mycorp.com", "secret-token")
		loader := NewTFCRemoteVariablesLoader(client, "", true, newDiscardLogger())

		vars, err := loader.Load(RemoteVarLoaderOptions{Blocks: blocks})
		require.NoError(t, err)
		assert.Empty(t, vars)
		// The client must not have been repointed at the untrusted host.
		assert.Equal(t, "tfe.mycorp.com", client.Host())
	})
}
