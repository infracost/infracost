package hcl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectLocator_FindRootModules_WithSingleProject(t *testing.T) {
	pl := NewProjectLocator(newDiscardLogger(), &ProjectLocatorConfig{})
	p := "./testdata/project_locator/single_project"
	mods := pl.FindRootModules(p)

	require.Len(t, mods, 1)
	assert.Contains(t, mods, RootPath{RepoPath: p, Path: "./testdata/project_locator/single_project", HasChanges: false})
}

func TestProjectLocator_FindRootModules_WithMultiProjectMixed(t *testing.T) {
	pl := NewProjectLocator(newDiscardLogger(), &ProjectLocatorConfig{})
	p := "./testdata/project_locator/multi_project_mixed"
	mods := pl.FindRootModules(p)

	require.Len(t, mods, 2)
	assert.Contains(t, mods, RootPath{RepoPath: p, Path: "testdata/project_locator/multi_project_mixed/with_provider"})
	assert.Contains(t, mods, RootPath{RepoPath: p, Path: "testdata/project_locator/multi_project_mixed/with_backend"})
}

func TestProjectLocator_FindRootModules_WithMultiProject(t *testing.T) {
	pl := NewProjectLocator(newDiscardLogger(), &ProjectLocatorConfig{})
	p := "./testdata/project_locator/multi_project_with_module"
	mods := pl.FindRootModules(p)

	require.Len(t, mods, 2)
	assert.Contains(t, mods, RootPath{RepoPath: p, Path: "testdata/project_locator/multi_project_with_module/dev"})
	assert.Contains(t, mods, RootPath{RepoPath: p, Path: "testdata/project_locator/multi_project_with_module/prod"})
}

func TestProjectLocator_FindRootModules_WithMultiProject_ExcludeDirs(t *testing.T) {
	pl := NewProjectLocator(newDiscardLogger(), &ProjectLocatorConfig{
		ExcludedDirs: []string{"dev"},
	})
	p := "./testdata/project_locator/multi_project_with_module"
	mods := pl.FindRootModules(p)

	require.Len(t, mods, 1)
	assert.Contains(t, mods, RootPath{RepoPath: p, Path: "testdata/project_locator/multi_project_with_module/prod"})
}

func TestProjectLocator_FindRootModules_WithMultiProject_WithObjectChanges(t *testing.T) {
	pl := NewProjectLocator(newDiscardLogger(), &ProjectLocatorConfig{
		ChangedObjects: []string{
			"./testdata/project_locator/multi_project_with_module/dev/main.tf",
		},
	})
	p := "./testdata/project_locator/multi_project_with_module"
	mods := pl.FindRootModules(p)

	require.Len(t, mods, 2)
	assert.Contains(t, mods, RootPath{RepoPath: p, Path: "testdata/project_locator/multi_project_with_module/dev", HasChanges: true})
	assert.Contains(t, mods, RootPath{RepoPath: p, Path: "testdata/project_locator/multi_project_with_module/prod", HasChanges: false})
}

func TestProjectLocator_FindRootModules_WithMultiProject_WithModuleObjectChanges(t *testing.T) {
	pl := NewProjectLocator(newDiscardLogger(), &ProjectLocatorConfig{
		ChangedObjects: []string{
			"./testdata/project_locator/multi_project_with_module/modules/example/main.tf",
		},
	})
	p := "./testdata/project_locator/multi_project_with_module"
	mods := pl.FindRootModules(p)

	require.Len(t, mods, 2)
	assert.Contains(t, mods, RootPath{RepoPath: p, Path: "testdata/project_locator/multi_project_with_module/dev", HasChanges: false})
	assert.Contains(t, mods, RootPath{RepoPath: p, Path: "testdata/project_locator/multi_project_with_module/prod", HasChanges: true})
}

func TestProjectLocator_FindRootModules_WithMultiProjectWithoutProviderBlocks(t *testing.T) {
	pl := NewProjectLocator(newDiscardLogger(), &ProjectLocatorConfig{
		UseAllPaths: true,
	})
	p := "./testdata/project_locator/multi_project_without_provider_blocks"
	mods := pl.FindRootModules(p)

	require.Len(t, mods, 3)
	assert.Contains(t, mods, RootPath{RepoPath: p, Path: "testdata/project_locator/multi_project_without_provider_blocks/dev"})
	assert.Contains(t, mods, RootPath{RepoPath: p, Path: "testdata/project_locator/multi_project_without_provider_blocks/prod"})
	assert.Contains(t, mods, RootPath{RepoPath: p, Path: "testdata/project_locator/multi_project_without_provider_blocks/modules/example"})
}

func TestProjectLocator_FindRootModules_WithMultiProjectWithoutProviderBlocks_ExludePaths(t *testing.T) {
	pl := NewProjectLocator(newDiscardLogger(), &ProjectLocatorConfig{
		UseAllPaths: true,
		ExcludedDirs: []string{
			"modules/**",
		},
	})
	p := "./testdata/project_locator/multi_project_without_provider_blocks"
	mods := pl.FindRootModules(p)

	require.Len(t, mods, 2)
	assert.Contains(t, mods, RootPath{RepoPath: p, Path: "testdata/project_locator/multi_project_without_provider_blocks/dev"})
	assert.Contains(t, mods, RootPath{RepoPath: p, Path: "testdata/project_locator/multi_project_without_provider_blocks/prod"})
}

func TestProjectLocator_FindRootModules_WithChildTfvarsFile(t *testing.T) {
	pl := NewProjectLocator(newDiscardLogger(), &ProjectLocatorConfig{})
	p := "./testdata/project_locator/multi_project_with_child_terraform_var_files"
	mods := pl.FindRootModules(p)

	require.Len(t, mods, 3)
	assert.Contains(t, mods, RootPath{
		RepoPath:         p,
		Path:             "testdata/project_locator/multi_project_with_child_terraform_var_files/app1",
		HasChildVarFiles: true,
		TerraformVarFiles: []string{
			"defaults.tfvars",
			"env/prod.tfvars",
			"env/test.tfvars",
		},
	})

	assert.Contains(t, mods, RootPath{
		RepoPath: p,
		Path:     "testdata/project_locator/multi_project_with_child_terraform_var_files/app1/app3",
		TerraformVarFiles: []string{
			"qa.tfvars",
		},
	})

	assert.Contains(t, mods, RootPath{
		RepoPath:         p,
		Path:             "testdata/project_locator/multi_project_with_child_terraform_var_files/app2",
		HasChildVarFiles: true,
		TerraformVarFiles: []string{
			"env/defaults.tfvars",
			"env/prod.tfvars",
			"env/staging.tfvars",
		},
	})
}
