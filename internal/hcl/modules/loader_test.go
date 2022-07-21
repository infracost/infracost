package modules

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/infracost/infracost/internal/credentials"
)

func testLoaderE2E(t *testing.T, path string, expectedModules []*ManifestModule, cleanup bool) {
	if cleanup {
		err := os.RemoveAll(filepath.Join(path, ".infracost"))
		assert.NoError(t, err)
	}

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	moduleLoader := NewModuleLoader(path, &CredentialsSource{FetchToken: credentials.FindTerraformCloudToken}, logrus.NewEntry(logger))

	manifest, err := moduleLoader.Load()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	sort.Slice(expectedModules, func(i, j int) bool {
		return expectedModules[i].Key < expectedModules[j].Key
	})

	actualModules := manifest.Modules

	sort.Slice(actualModules, func(i, j int) bool {
		return actualModules[i].Key < actualModules[j].Key
	})

	assert.Equal(t, expectedModules, actualModules)
}

func TestNestedModules(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	testLoaderE2E(t, "./testdata/nested_modules", []*ManifestModule{
		{
			Key:    "local-module",
			Source: "./modules/local-module",
			Dir:    "modules/local-module",
		},
		{
			Key:     "registry-module",
			Source:  "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version: "3.4.0",
			Dir:     ".infracost/terraform_modules/registry-module",
		},
		{
			Key:    "git-module",
			Source: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git",
			Dir:    ".infracost/terraform_modules/git-module",
		},
		{
			Key:    "local-module.nested-local-module",
			Source: "./nested-local-module",
			Dir:    "modules/local-module/nested-local-module",
		},
		{
			Key:     "local-module.nested-registry-module",
			Source:  "registry.terraform.io/terraform-aws-modules/sns/aws",
			Version: "3.1.0",
			Dir:     ".infracost/terraform_modules/local-module.nested-registry-module",
		},
		{
			Key:    "local-module.nested-git-module",
			Source: "git::https://github.com/terraform-aws-modules/terraform-aws-sns.git",
			Dir:    ".infracost/terraform_modules/local-module.nested-git-module",
		},
	}, true)
}

func TestSubmodules(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	testLoaderE2E(t, "./testdata/submodules", []*ManifestModule{
		{
			Key:     "registry-submodule",
			Source:  "registry.terraform.io/terraform-aws-modules/route53/aws//modules/zones",
			Version: "2.5.0",
			Dir:     ".infracost/terraform_modules/registry-submodule/modules/zones",
		},
		{
			Key:    "git-submodule",
			Source: "git::https://github.com/terraform-aws-modules/terraform-aws-route53.git//modules/zones",
			Dir:    ".infracost/terraform_modules/git-submodule/modules/zones",
		},
	}, true)
}

func TestModuleMultipleUses(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	testLoaderE2E(t, "./testdata/module_multiple_uses", []*ManifestModule{
		{
			Key:     "registry-module-1",
			Source:  "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version: "3.4.0",
			Dir:     ".infracost/terraform_modules/registry-module-1",
		},
		{
			Key:     "registry-module-2",
			Source:  "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version: "3.4.0",
			Dir:     ".infracost/terraform_modules/registry-module-2",
		},
	}, true)
}

func TestModuleMultipleUsesMissingManifest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	expectedModules := []*ManifestModule{
		{
			Key:     "registry-module-1",
			Source:  "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version: "3.4.0",
			Dir:     ".infracost/terraform_modules/registry-module-1",
		},
		{
			Key:     "registry-module-2",
			Source:  "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version: "3.4.0",
			Dir:     ".infracost/terraform_modules/registry-module-2",
		},
	}

	// Run first time to download modules
	testLoaderE2E(t, "./testdata/module_multiple_uses", expectedModules, true)

	// Remove the manifest file to test we can still work with broken manifests
	err := os.Remove("./testdata/module_multiple_uses/.infracost/terraform_modules/manifest.json")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// Re-run without cleaning up the modules directory
	testLoaderE2E(t, "./testdata/module_multiple_uses", expectedModules, false)
}

func TestWithCachedModules(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	testLoaderE2E(t, "./testdata/with_cached_modules", []*ManifestModule{
		{
			Key:    "local-module",
			Source: "./modules/local-module",
			Dir:    "modules/local-module",
		},
		{
			Key:     "registry-module",
			Source:  "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version: "3.4.0",
			Dir:     ".infracost/terraform_modules/registry-module",
		},
		{
			Key:    "git-module",
			Source: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git",
			Dir:    ".infracost/terraform_modules/git-module",
		},
	}, false)

	// Check that the modules were not overwritten
	regModContents, err := os.ReadFile("./testdata/with_cached_modules/.infracost/terraform_modules/registry-module/main.tf")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	gitModContents, err := os.ReadFile("./testdata/with_cached_modules/.infracost/terraform_modules/git-module/main.tf")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, string(regModContents), "// Placeholder file\n")
	assert.Equal(t, string(gitModContents), "// Placeholder file\n")
}
