package modules

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/credentials"
	sync2 "github.com/infracost/infracost/internal/sync"
)

type TestLoaderE2EOpts = struct {
	SourceMap config.TerraformSourceMap
	Cleanup   bool
	IgnoreDir bool
}

func testLoaderE2E(t *testing.T, path string, expectedModules []*ManifestModule, opts TestLoaderE2EOpts) {
	if opts.Cleanup {
		err := os.RemoveAll(filepath.Join(path, config.InfracostDir))
		assert.NoError(t, err)
	}

	logger := zerolog.New(io.Discard)

	moduleLoader := NewModuleLoader(path, NewSharedHCLParser(), &CredentialsSource{FetchToken: credentials.FindTerraformCloudToken}, opts.SourceMap, logger, &sync2.KeyMutex{})

	manifest, err := moduleLoader.Load(path)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	sort.Slice(expectedModules, func(i, j int) bool {
		return expectedModules[i].Key < expectedModules[j].Key
	})

	expected := []*ManifestModule{}
	for _, m := range expectedModules {
		e := &ManifestModule{
			Key:         m.Key,
			Source:      m.Source,
			Version:     m.Version,
			DownloadURL: m.DownloadURL,
		}

		if !opts.IgnoreDir {
			e.Dir = m.Dir
		}

		expected = append(expected, e)
	}

	actualModules := manifest.Modules

	sort.Slice(actualModules, func(i, j int) bool {
		return actualModules[i].Key < actualModules[j].Key
	})

	if opts.IgnoreDir {
		for _, m := range actualModules {
			m.Dir = ""
		}
	}

	assert.Equal(t, expected, actualModules)
}

func TestNestedModules(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	testLoaderE2E(t, "./testdata/nested_modules", []*ManifestModule{
		{
			Key:         "registry-module",
			Source:      "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version:     "3.4.0",
			Dir:         ".infracost/terraform_modules/f8b5f5ddb85ee755b31c8b76d2801f5b",
			DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance?ref=v3.4.0",
		},
		{
			Key:         "registry-module-different-name",
			Source:      "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version:     "3.4.0",
			Dir:         ".infracost/terraform_modules/f8b5f5ddb85ee755b31c8b76d2801f5b",
			DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance?ref=v3.4.0",
		},
		{
			Key:    "git-module",
			Source: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git",
			Dir:    ".infracost/terraform_modules/9740179dc58fea6ce4a32fdc5b4e0839",
		},
		{
			Key:    "git-module-different-name",
			Source: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git",
			Dir:    ".infracost/terraform_modules/9740179dc58fea6ce4a32fdc5b4e0839",
		},
		{
			Key:         "local-module.nested-registry-module",
			Source:      "registry.terraform.io/terraform-aws-modules/sns/aws",
			Version:     "3.1.0",
			Dir:         ".infracost/terraform_modules/b72552bcaa63a49783f9a735a697c662",
			DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-sns?ref=v3.1.0",
		},
		{
			Key:    "local-module.nested-git-module",
			Source: "git::https://github.com/terraform-aws-modules/terraform-aws-sns.git",
			Dir:    ".infracost/terraform_modules/db69103dcf4b9586b710a97de31750bd",
		},
		{
			Key:         "local-module.nested-registry-module-using-same-source",
			Source:      "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version:     "3.4.0",
			Dir:         ".infracost/terraform_modules/f8b5f5ddb85ee755b31c8b76d2801f5b",
			DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance?ref=v3.4.0",
		},
	}, TestLoaderE2EOpts{Cleanup: true})
}

func TestSubmodules(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	testLoaderE2E(t, "./testdata/submodules", []*ManifestModule{
		{
			Key:         "registry-submodule",
			Source:      "registry.terraform.io/terraform-aws-modules/route53/aws//modules/zones",
			Version:     "2.5.0",
			Dir:         ".infracost/terraform_modules/d1e3bab8b33f57431ace737ccffbf67f/modules/zones",
			DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-route53?ref=v2.5.0",
		},
		{
			Key:         "registry-submodule-records",
			Source:      "registry.terraform.io/terraform-aws-modules/route53/aws//modules/records",
			Version:     "2.5.0",
			Dir:         ".infracost/terraform_modules/d1e3bab8b33f57431ace737ccffbf67f/modules/records",
			DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-route53?ref=v2.5.0",
		},
		{
			Key:    "git-submodule",
			Source: "git::https://github.com/terraform-aws-modules/terraform-aws-route53.git//modules/zones",
			Dir:    ".infracost/terraform_modules/03c49f2fce2b8552355561b7ac4f2c94/modules/zones",
		},
	}, TestLoaderE2EOpts{Cleanup: true})
}

func TestModuleMultipleUses(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	testLoaderE2E(t, "./testdata/module_multiple_uses", []*ManifestModule{
		{
			Key:         "registry-module-1",
			Source:      "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version:     "3.4.0",
			Dir:         ".infracost/terraform_modules/f8b5f5ddb85ee755b31c8b76d2801f5b",
			DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance?ref=v3.4.0",
		},
		{
			Key:         "registry-module-2",
			Source:      "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version:     "3.4.0",
			Dir:         ".infracost/terraform_modules/f8b5f5ddb85ee755b31c8b76d2801f5b",
			DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance?ref=v3.4.0",
		},
	}, TestLoaderE2EOpts{Cleanup: true})
}

func TestModuleMultipleUsesMissingManifest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	expectedModules := []*ManifestModule{
		{
			Key:         "registry-module-1",
			Source:      "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version:     "3.4.0",
			Dir:         ".infracost/terraform_modules/f8b5f5ddb85ee755b31c8b76d2801f5b",
			DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance?ref=v3.4.0",
		},
		{
			Key:         "registry-module-2",
			Source:      "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version:     "3.4.0",
			Dir:         ".infracost/terraform_modules/f8b5f5ddb85ee755b31c8b76d2801f5b",
			DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance?ref=v3.4.0",
		},
	}

	// Run first time to download modules
	testLoaderE2E(t, "./testdata/module_multiple_uses", expectedModules, TestLoaderE2EOpts{Cleanup: false})

	// Remove the manifest file to test we can still work with broken manifests
	err := os.Remove("./testdata/module_multiple_uses/.infracost/terraform_modules/manifest.json")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// Re-run without cleaning up the modules directory
	testLoaderE2E(t, "./testdata/module_multiple_uses", expectedModules, TestLoaderE2EOpts{Cleanup: false})
}

func TestWithCachedModules(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	testLoaderE2E(t, "./testdata/with_cached_modules", []*ManifestModule{
		{
			Key:         "registry-module",
			Source:      "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version:     "3.4.0",
			Dir:         ".infracost/terraform_modules/f8b5f5ddb85ee755b31c8b76d2801f5b",
			DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance?ref=v3.4.0",
		},
		{
			Key:    "git-module",
			Source: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git",
			Dir:    ".infracost/terraform_modules/9740179dc58fea6ce4a32fdc5b4e0839",
		},
	}, TestLoaderE2EOpts{Cleanup: false})

	// Check that the modules were not overwritten
	regModContents, err := os.ReadFile("./testdata/with_cached_modules/.infracost/terraform_modules/f8b5f5ddb85ee755b31c8b76d2801f5b/main.tf")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	gitModContents, err := os.ReadFile("./testdata/with_cached_modules/.infracost/terraform_modules/9740179dc58fea6ce4a32fdc5b4e0839/main.tf")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, string(regModContents), "// registry-module\n")
	assert.Equal(t, string(gitModContents), "// git-module\n")
}

func TestMultiProject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	path := "./testdata/multi_project"
	err := os.RemoveAll(filepath.Join(path, config.InfracostDir))
	assert.NoError(t, err)

	logger := zerolog.New(io.Discard)

	moduleLoader := NewModuleLoader(path, NewSharedHCLParser(), &CredentialsSource{FetchToken: credentials.FindTerraformCloudToken}, config.TerraformSourceMap{}, logger, &sync2.KeyMutex{})

	wg := &sync.WaitGroup{}
	wg.Add(3)
	go func(t *testing.T) {
		t.Helper()
		_, err := moduleLoader.Load(filepath.Join(path, "dev"))
		wg.Done()
		assert.NoError(t, err)
	}(t)

	go func(t *testing.T) {
		t.Helper()
		_, err := moduleLoader.Load(filepath.Join(path, "prod"))
		wg.Done()
		assert.NoError(t, err)
	}(t)

	go func(t *testing.T) {
		t.Helper()
		_, err := moduleLoader.Load(filepath.Join(path, "with_existing_terraform_mods"))
		wg.Done()
		assert.NoError(t, err)
	}(t)

	wg.Wait()
	assertModulesEqual(t, moduleLoader, filepath.Join(path, "prod"), []*ManifestModule{
		{
			Key:         "registry-module",
			Source:      "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version:     "3.4.0",
			Dir:         ".infracost/terraform_modules/f8b5f5ddb85ee755b31c8b76d2801f5b",
			DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance?ref=v3.4.0",
		},
		{
			Key:         "registry-module-different-name",
			Source:      "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version:     "3.4.0",
			Dir:         ".infracost/terraform_modules/f8b5f5ddb85ee755b31c8b76d2801f5b",
			DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance?ref=v3.4.0",
		},
		{
			Key:    "git-module",
			Source: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git",
			Dir:    ".infracost/terraform_modules/9740179dc58fea6ce4a32fdc5b4e0839",
		},
		{
			Key:    "git-module-different-name",
			Source: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git",
			Dir:    ".infracost/terraform_modules/9740179dc58fea6ce4a32fdc5b4e0839",
		},
		{
			Key:         "local-module.nested-registry-module",
			Source:      "registry.terraform.io/terraform-aws-modules/sns/aws",
			Version:     "3.1.0",
			Dir:         ".infracost/terraform_modules/b72552bcaa63a49783f9a735a697c662",
			DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-sns?ref=v3.1.0",
		},
		{
			Key:    "local-module.nested-git-module",
			Source: "git::https://github.com/terraform-aws-modules/terraform-aws-sns.git",
			Dir:    ".infracost/terraform_modules/db69103dcf4b9586b710a97de31750bd",
		},
		{
			Key:         "local-module.nested-registry-module-using-same-source",
			Source:      "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version:     "3.4.0",
			Dir:         ".infracost/terraform_modules/f8b5f5ddb85ee755b31c8b76d2801f5b",
			DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance?ref=v3.4.0",
		},
	})

	assertModulesEqual(t, moduleLoader, filepath.Join(path, "dev"), []*ManifestModule{
		{
			Key:         "registry-module",
			Source:      "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version:     "3.4.0",
			Dir:         ".infracost/terraform_modules/f8b5f5ddb85ee755b31c8b76d2801f5b",
			DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance?ref=v3.4.0",
		},
		{
			Key:         "registry-module-different-name",
			Source:      "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version:     "3.4.0",
			Dir:         ".infracost/terraform_modules/f8b5f5ddb85ee755b31c8b76d2801f5b",
			DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance?ref=v3.4.0",
		},
		{
			Key:    "git-module",
			Source: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git",
			Dir:    ".infracost/terraform_modules/9740179dc58fea6ce4a32fdc5b4e0839",
		},
		{
			Key:    "git-module-different-name",
			Source: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git",
			Dir:    ".infracost/terraform_modules/9740179dc58fea6ce4a32fdc5b4e0839",
		},
		{
			Key:    "another-git-module-only-in-dev",
			Source: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git",
			Dir:    ".infracost/terraform_modules/9740179dc58fea6ce4a32fdc5b4e0839",
		},
		{
			Key:         "local-module.nested-registry-module",
			Source:      "registry.terraform.io/terraform-aws-modules/sns/aws",
			Version:     "3.1.0",
			Dir:         ".infracost/terraform_modules/b72552bcaa63a49783f9a735a697c662",
			DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-sns?ref=v3.1.0",
		},
		{
			Key:    "local-module.nested-git-module",
			Source: "git::https://github.com/terraform-aws-modules/terraform-aws-sns.git",
			Dir:    ".infracost/terraform_modules/db69103dcf4b9586b710a97de31750bd",
		},
		{
			Key:         "local-module.nested-registry-module-using-same-source",
			Source:      "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version:     "3.4.0",
			Dir:         ".infracost/terraform_modules/f8b5f5ddb85ee755b31c8b76d2801f5b",
			DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance?ref=v3.4.0",
		},
	})

	assertModulesEqual(t, moduleLoader, filepath.Join(path, "with_existing_terraform_mods"), []*ManifestModule{
		{
			Key:    "",
			Source: "",
			Dir:    "with_existing_terraform_mods",
		},
		{
			Key:    "local-module",
			Source: "../modules/local-module",
			Dir:    "modules/local-module",
		},
		{
			Key:    "local-module.nested-git-module",
			Source: "git::https://github.com/terraform-aws-modules/terraform-aws-sns.git",
			Dir:    "with_existing_terraform_mods/.terraform/modules/local-module.nested-git-module",
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
			Dir:     "with_existing_terraform_mods/.terraform/modules/local-module.nested-registry-module",
		},
		{
			Key:     "local-module.nested-registry-module-using-same-source",
			Source:  "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version: "3.4.0",
			Dir:     "with_existing_terraform_mods/.terraform/modules/local-module.nested-registry-module-using-same-source",
		},
		{
			Key:     "registry-module",
			Source:  "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			Version: "3.4.0",
			Dir:     "with_existing_terraform_mods/.terraform/modules/registry-module",
		},
	})
}

func TestSourceMapRegistryModule(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	tests := []struct {
		name      string
		sourceMap config.TerraformSourceMap
		want      []*ManifestModule
	}{
		{
			name:      "empty source map",
			sourceMap: config.TerraformSourceMap{},
			want: []*ManifestModule{
				{
					Key:         "registry-module",
					Source:      "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
					Version:     "3.4.0",
					DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance?ref=v3.4.0",
				},
			},
		},
		{
			name: "git without version",
			sourceMap: config.TerraformSourceMap{
				"terraform-aws-modules/ec2-instance/aws": "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git",
			},
			want: []*ManifestModule{
				{
					Key:    "registry-module",
					Source: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git",
					// We can't map the version here when going from registry to git, since we don't know which tag it should be.
					// Some modules are prefixed with v and some aren't.
					Version: "",
				},
			},
		},
		{
			name: "git with version",
			sourceMap: config.TerraformSourceMap{
				"terraform-aws-modules/ec2-instance/aws": "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git?ref=v5.5.0",
			},
			want: []*ManifestModule{
				{
					Key:     "registry-module",
					Source:  "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git?ref=v5.5.0",
					Version: "",
				},
			},
		},
		// This shouldn't map since the source in the code doesn't include the registry hostname
		{
			name: "with registry hostname",
			sourceMap: config.TerraformSourceMap{
				"registry.terraform.io/terraform-aws-modules/ec2-instance/aws": "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git",
			},
			want: []*ManifestModule{
				{
					Key:         "registry-module",
					Source:      "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
					Version:     "3.4.0",
					DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance?ref=v3.4.0",
				},
			},
		},
		{
			name: "with prefix",
			sourceMap: config.TerraformSourceMap{
				"terraform-aws-modules/": "registry.terraform.io/terraform-aws-modules/",
			},
			want: []*ManifestModule{
				{
					Key:         "registry-module",
					Source:      "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
					Version:     "3.4.0",
					DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance?ref=v3.4.0",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLoaderE2E(t, "./testdata/sourcemap_registry_module", tt.want, TestLoaderE2EOpts{SourceMap: tt.sourceMap, Cleanup: true, IgnoreDir: true})
		})
	}
}

func TestSourceMapRegistrySubmodule(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	tests := []struct {
		name      string
		sourceMap config.TerraformSourceMap
		want      []*ManifestModule
	}{
		{
			name:      "empty source map",
			sourceMap: config.TerraformSourceMap{},
			want: []*ManifestModule{
				{
					Key:         "registry-submodule",
					Source:      "registry.terraform.io/terraform-aws-modules/route53/aws//modules/zones",
					Version:     "2.5.0",
					DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-route53?ref=v2.5.0",
				},
			},
		},
		{
			name: "git without version",
			sourceMap: config.TerraformSourceMap{
				"terraform-aws-modules/route53/aws": "git::https://github.com/terraform-aws-modules/terraform-aws-route53.git",
			},
			want: []*ManifestModule{
				{
					Key:    "registry-submodule",
					Source: "git::https://github.com/terraform-aws-modules/terraform-aws-route53.git//modules/zones",
					// We can't map the version here when going from registry to git, since we don't know which tag it should be.
					// Some modules are prefixed with v and some aren't.
					Version: "",
				},
			},
		},
		{
			name: "git with version",
			sourceMap: config.TerraformSourceMap{
				"terraform-aws-modules/route53/aws": "git::https://github.com/terraform-aws-modules/terraform-aws-route53.git?ref=v2.10.2",
			},
			want: []*ManifestModule{
				{
					Key:     "registry-submodule",
					Source:  "git::https://github.com/terraform-aws-modules/terraform-aws-route53.git//modules/zones?ref=v2.10.2",
					Version: "",
				},
			},
		},
		// This shouldn't map since the source in the code doesn't include the registry hostname
		{
			name: "with registry hostname",
			sourceMap: config.TerraformSourceMap{
				"registry.terraform.io/terraform-aws-modules/route53/aws": "git::https://github.com/terraform-aws-modules/terraform-aws-route53.git",
			},
			want: []*ManifestModule{
				{
					Key:         "registry-submodule",
					Source:      "registry.terraform.io/terraform-aws-modules/route53/aws//modules/zones",
					Version:     "2.5.0",
					DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-route53?ref=v2.5.0",
				},
			},
		},
		{
			name: "with prefix",
			sourceMap: config.TerraformSourceMap{
				"terraform-aws-modules/": "registry.terraform.io/terraform-aws-modules/",
			},
			want: []*ManifestModule{
				{
					Key:         "registry-submodule",
					Source:      "registry.terraform.io/terraform-aws-modules/route53/aws//modules/zones",
					Version:     "2.5.0",
					DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-route53?ref=v2.5.0",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLoaderE2E(t, "./testdata/sourcemap_registry_submodule", tt.want, TestLoaderE2EOpts{SourceMap: tt.sourceMap, Cleanup: true, IgnoreDir: true})
		})
	}
}

func TestSourceMapGitModule(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	tests := []struct {
		name      string
		sourceMap config.TerraformSourceMap
		want      []*ManifestModule
	}{
		{
			name:      "empty source map",
			sourceMap: config.TerraformSourceMap{},
			want: []*ManifestModule{
				{
					Key:     "git-module",
					Source:  "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git?ref=v3.4.0",
					Version: "",
				},
			},
		},
		{
			name: "git without version",
			sourceMap: config.TerraformSourceMap{
				"git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git": "git::https://github.com/infracost-tests/terraform-aws-ec2-instance.git",
			},
			want: []*ManifestModule{
				{
					Key:     "git-module",
					Source:  "git::https://github.com/infracost-tests/terraform-aws-ec2-instance.git?ref=v3.4.0",
					Version: "",
				},
			},
		},
		{
			name: "git with prefix",
			sourceMap: config.TerraformSourceMap{
				"git::https://github.com/terraform-aws-modules/": "git::https://github.com/infracost-tests/",
			},
			want: []*ManifestModule{
				{
					Key:     "git-module",
					Source:  "git::https://github.com/infracost-tests/terraform-aws-ec2-instance.git?ref=v3.4.0",
					Version: "",
				},
			},
		},
		{
			name: "git with ref override",
			sourceMap: config.TerraformSourceMap{
				"git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git?ref=v3.4.0": "git::https://github.com/infracost-tests/terraform-aws-ec2-instance.git?ref=master",
				"git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git":            "git::https://github.com/infracost-tests/terraform-aws-ec2-instance.git",
			},
			want: []*ManifestModule{
				{
					Key:     "git-module",
					Source:  "git::https://github.com/infracost-tests/terraform-aws-ec2-instance.git?ref=master",
					Version: "",
				},
			},
		},
		{
			name: "git with repo override",
			sourceMap: config.TerraformSourceMap{
				"git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git": "git::https://github.com/infracost-tests/terraform-aws-ec2-instance.git",
				"git::https://github.com/terraform-aws-modules/":                               "git::https://github.com/infracost-tests/",
			},
			want: []*ManifestModule{
				{
					Key:     "git-module",
					Source:  "git::https://github.com/infracost-tests/terraform-aws-ec2-instance.git?ref=v3.4.0",
					Version: "",
				},
			},
		},
		{
			name: "git to registry",
			sourceMap: config.TerraformSourceMap{
				"git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance.git": "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
			},
			want: []*ManifestModule{
				{
					Key:         "git-module",
					Source:      "registry.terraform.io/terraform-aws-modules/ec2-instance/aws",
					Version:     "3.4.0",
					DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-ec2-instance?ref=v3.4.0",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLoaderE2E(t, "./testdata/sourcemap_git_module", tt.want, TestLoaderE2EOpts{SourceMap: tt.sourceMap, Cleanup: true, IgnoreDir: true})
		})
	}
}

func TestSourceMapGitSubmodule(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	tests := []struct {
		name      string
		sourceMap config.TerraformSourceMap
		want      []*ManifestModule
	}{
		{
			name:      "empty source map",
			sourceMap: config.TerraformSourceMap{},
			want: []*ManifestModule{
				{
					Key:     "git-submodule",
					Source:  "git::https://github.com/terraform-aws-modules/terraform-aws-route53.git//modules/zones?ref=v2.5.0",
					Version: "",
				},
			},
		},
		{
			name: "git without version",
			sourceMap: config.TerraformSourceMap{
				"git::https://github.com/terraform-aws-modules/terraform-aws-route53.git": "git::https://github.com/infracost-tests/terraform-aws-route53.git",
			},
			want: []*ManifestModule{
				{
					Key:     "git-submodule",
					Source:  "git::https://github.com/infracost-tests/terraform-aws-route53.git//modules/zones?ref=v2.5.0",
					Version: "",
				},
			},
		},
		{
			name: "git with prefix",
			sourceMap: config.TerraformSourceMap{
				"git::https://github.com/terraform-aws-modules/": "git::https://github.com/infracost-tests/",
			},
			want: []*ManifestModule{
				{
					Key:     "git-submodule",
					Source:  "git::https://github.com/infracost-tests/terraform-aws-route53.git//modules/zones?ref=v2.5.0",
					Version: "",
				},
			},
		},
		{
			name: "git with ref override",
			sourceMap: config.TerraformSourceMap{
				"git::https://github.com/terraform-aws-modules/terraform-aws-route53.git?ref=v2.5.0": "git::https://github.com/infracost-tests/terraform-aws-route53.git?ref=master",
				"git::https://github.com/terraform-aws-modules/terraform-aws-route53.git":            "git::https://github.com/infracost-tests/terraform-aws-route53.git",
			},
			want: []*ManifestModule{
				{
					Key:     "git-submodule",
					Source:  "git::https://github.com/infracost-tests/terraform-aws-route53.git//modules/zones?ref=master",
					Version: "",
				},
			},
		},
		{
			name: "git with repo override",
			sourceMap: config.TerraformSourceMap{
				"git::https://github.com/terraform-aws-modules/terraform-aws-route53.git": "git::https://github.com/infracost-tests/terraform-aws-route53.git",
				"git::https://github.com/terraform-aws-modules/":                          "git::https://github.com/infracost-tests/",
			},
			want: []*ManifestModule{
				{
					Key:     "git-submodule",
					Source:  "git::https://github.com/infracost-tests/terraform-aws-route53.git//modules/zones?ref=v2.5.0",
					Version: "",
				},
			},
		},
		{
			name: "git to registry",
			sourceMap: config.TerraformSourceMap{
				"git::https://github.com/terraform-aws-modules/terraform-aws-route53.git": "registry.terraform.io/terraform-aws-modules/route53/aws",
			},
			want: []*ManifestModule{
				{
					Key:         "git-submodule",
					Source:      "registry.terraform.io/terraform-aws-modules/route53/aws//modules/zones",
					Version:     "2.5.0",
					DownloadURL: "git::https://github.com/terraform-aws-modules/terraform-aws-route53?ref=v2.5.0",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLoaderE2E(t, "./testdata/sourcemap_git_submodule", tt.want, TestLoaderE2EOpts{SourceMap: tt.sourceMap, Cleanup: true, IgnoreDir: true})
		})
	}
}

func assertModulesEqual(t *testing.T, moduleLoader *ModuleLoader, path string, expectedModules []*ManifestModule) {
	t.Helper()

	manifest, err := moduleLoader.Load(path)
	assert.NoError(t, err)
	actualModules := manifest.Modules

	sort.Slice(expectedModules, func(i, j int) bool {
		return expectedModules[i].Key < expectedModules[j].Key
	})

	sort.Slice(actualModules, func(i, j int) bool {
		return actualModules[i].Key < actualModules[j].Key
	})

	assert.Equal(t, expectedModules, actualModules)
}
