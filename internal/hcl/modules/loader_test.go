package modules

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNestedModules(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	path := "./testdata/nested_modules"
	err := os.RemoveAll(filepath.Join(path, ".infracost"))
	assert.NoError(t, err)

	moduleLoader := &ModuleLoader{
		Path: path,
	}

	manifest, err := moduleLoader.Load()
	assert.NoError(t, err)

	expectedModules := []*ManifestModule{
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
