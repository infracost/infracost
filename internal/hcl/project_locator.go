package hcl

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gobwas/glob"
	tgconfig "github.com/gruntwork-io/terragrunt/config"
	"github.com/gruntwork-io/terragrunt/options"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/rs/zerolog"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/infracost/infracost/internal/config"
)

var (
	defaultEnvs = []string{
		"prd",
		"prod",
		"production",
		"preprod",
		"staging",
		"stage",
		"stg",
		"stag",
		"development",
		"dev",
		"release",
		"testing",
		"test",
		"tst",
		"qa",
		"uat",
		"live",
		"sandbox",
		"sbx",
		"sbox",
		"demo",
		"integration",
		"int",
		"experimental",
		"experiments",
		"trial",
		"validation",
		"perf",
		"sec",
		"dr",
		"load",
		"management",
		"mgmt",
		"playground",
	}
	defaultExtensions = []string{
		".tfvars",
		".auto.tfvars",
		".tfvars.json",
		".auto.tfvars.json",
	}
)

// EnvFileMatcher is used to match environment specific var files.
type EnvFileMatcher struct {
	envNames   []string
	envLookup  map[string]struct{}
	extensions []string
	wildcards  []string
}

func CreateEnvFileMatcher(names []string, extensions []string) *EnvFileMatcher {
	if extensions == nil {
		// create a matcher with the .json extensions as well so that we support
		// tfars-env.json use case.
		extensions = append(defaultExtensions, ".json") // nolint
	}

	if len(names) == 0 {
		return CreateEnvFileMatcher(defaultEnvs, extensions)
	}

	// Sort the extensions by length so that we always prefer the longest extension
	// when matching a file.
	sort.Slice(extensions, func(i, j int) bool {
		return len(extensions[i]) > len(extensions[j])
	})

	var envNames []string
	var wildcards []string
	for _, name := range names {
		// envNames can contain wildcards, we need to handle them separately. e.g: dev-*
		// will create separate envs for dev-staging and dev-legacy. We don't want these
		// wildcards to appear in the envNames list as this will create unwanted env
		// grouping.
		if strings.Contains(name, "*") {
			wildcards = append(wildcards, name)
			continue
		}

		// ensure all env names to lowercase, so we can match case insensitively.
		envNames = append(envNames, strings.ToLower(name))
	}

	lookup := make(map[string]struct{}, len(names))
	for _, name := range envNames {
		lookup[name] = struct{}{}
	}

	return &EnvFileMatcher{
		envNames:   envNames,
		envLookup:  lookup,
		extensions: extensions,
		wildcards:  wildcards,
	}
}

// IsAutoVarFile checks if the var file is an auto.tfvars or terraform.tfvars.
// These are special Terraform var files that are applied to every project
// automatically.
func IsAutoVarFile(file string) bool {
	withoutJSONSuffix := strings.TrimSuffix(file, ".json")

	return strings.HasSuffix(withoutJSONSuffix, ".auto.tfvars") || withoutJSONSuffix == "terraform.tfvars"
}

// IsGlobalVarFile checks if the var file is a global var file.
func (e *EnvFileMatcher) IsGlobalVarFile(file string) bool {
	return !e.IsEnvName(file)
}

// IsEnvName checks if the var file is an environment specific var file.
func (e *EnvFileMatcher) IsEnvName(file string) bool {
	clean := e.clean(file)
	_, ok := e.envLookup[clean]
	if ok {
		return true
	}

	for _, name := range e.envNames {
		if e.hasEnvPrefix(clean, name) || e.hasEnvSuffix(clean, name) {
			return true
		}
	}

	for _, wildcard := range e.wildcards {
		if isMatch, _ := path.Match(wildcard, clean); isMatch {
			return true
		}
	}

	return false
}

func (e *EnvFileMatcher) clean(name string) string {
	base := filepath.Base(name)

	stem, _ := splitVarFileExt(base, e.extensions)
	return strings.ToLower(stem)
}

// EnvName returns the environment name for the given var file.
func (e *EnvFileMatcher) EnvName(file string) string {
	// if we have a direct match to an env name, return it.
	clean := e.clean(file)
	_, ok := e.envLookup[clean]
	if ok {
		return clean
	}

	// if we have a wildcard match to an env name return the clean name now
	// as the partial match logic can collide with wildcard matches.
	for _, wildcard := range e.wildcards {
		if isMatch, _ := path.Match(wildcard, clean); isMatch {
			return clean
		}
	}

	// if we have a partial suffix match to an env name return the partial match
	// which is the longest match. This is likely to be the better match. e.g: if we
	// have both dev and legacy-dev as defined envNames, given a tfvar named
	// legacy-dev-staging legacy-dev should be the env name returned.
	var match string
	for _, name := range e.envNames {
		if e.hasEnvSuffix(clean, name) {
			if len(name) > len(match) {
				match = name
			}
		}
	}

	if match != "" {
		return match
	}

	// repeat the same process for suffixes but with prefix matches.
	for _, name := range e.envNames {
		if e.hasEnvPrefix(clean, name) {
			if len(name) > len(match) {
				match = name
			}
		}
	}

	if match != "" {
		return match
	}

	return clean
}

func (e *EnvFileMatcher) hasEnvPrefix(clean string, name string) bool {
	return strings.HasPrefix(clean, name+"-") || strings.HasPrefix(clean, name+"_") || strings.HasPrefix(clean, "."+name)
}

func (e *EnvFileMatcher) hasEnvSuffix(clean string, name string) bool {
	return strings.HasSuffix(clean, "_"+name) || strings.HasSuffix(clean, "-"+name) || strings.HasSuffix(clean, "."+name)
}

type discoveredProject struct {
	isTerragrunt bool

	hasProviderBlock bool
	hasBackendBlock  bool
	depth            int
	files            map[string]*hcl.File
	path             string
}

func (p discoveredProject) hasRootModuleBlocks() bool {
	return p.hasBackendBlock || p.hasProviderBlock
}

// ProjectLocator finds Terraform projects for given paths.
// It naively excludes folders that are imported as modules in other projects.
type ProjectLocator struct {
	modules           map[string]struct{}
	projectDuplicates map[string]bool
	moduleCalls       map[string][]string
	excludedDirs      []string
	changedObjects    []string
	useAllPaths       bool
	logger            zerolog.Logger

	basePath           string
	discoveredVarFiles map[string][]RootPathVarFile
	discoveredProjects []discoveredProject
	includedDirs       []string
	envMatcher         *EnvFileMatcher
	skip               bool

	shouldSkipDir              func(string) bool
	shouldIncludeDir           func(string) bool
	pathOverrides              []pathOverride
	wdContainsTerragrunt       bool
	fallbackToIncludePaths     bool
	maxConfiguredDepth         int
	forceProjectType           string
	terraformVarFileExtensions []string
	hclParser                  *hclparse.Parser
	hasCustomEnvExt            bool
	workingDirectory           string
}

// ProjectLocatorConfig provides configuration options on how the locator functions.
type ProjectLocatorConfig struct {
	ExcludedDirs               []string
	ChangedObjects             []string
	UseAllPaths                bool
	SkipAutoDetection          bool
	IncludedDirs               []string
	EnvNames                   []string
	PathOverrides              []PathOverrideConfig
	FallbackToIncludePaths     bool
	MaxSearchDepth             int
	ForceProjectType           string
	TerraformVarFileExtensions []string
	WorkingDirectory           string
}

type PathOverrideConfig struct {
	Path    string
	Exclude []string
	Only    []string
}

type pathOverride struct {
	glob glob.Glob

	exclude map[string]struct{}
	only    map[string]struct{}
}

func newPathOverride(override PathOverrideConfig) pathOverride {
	exclude := make(map[string]struct{}, len(override.Exclude))
	for _, s := range override.Exclude {
		exclude[s] = struct{}{}
	}
	only := make(map[string]struct{}, len(override.Only))
	for _, s := range override.Only {
		only[s] = struct{}{}
	}

	return pathOverride{
		glob:    glob.MustCompile(override.Path),
		exclude: exclude,
		only:    only,
	}
}

// NewProjectLocator returns safely initialized ProjectLocator.
func NewProjectLocator(logger zerolog.Logger, config *ProjectLocatorConfig) *ProjectLocator {
	matcher := CreateEnvFileMatcher(nil, nil)
	if config != nil {
		extensions := defaultExtensions
		if config.TerraformVarFileExtensions != nil {
			extensions = config.TerraformVarFileExtensions
			matcher = CreateEnvFileMatcher(config.EnvNames, config.TerraformVarFileExtensions)
		} else {
			matcher = CreateEnvFileMatcher(config.EnvNames, nil)
		}

		// Sort the extensions by length so that we always prefer the longest extension
		// when matching a file.
		sort.Slice(extensions, func(i, j int) bool {
			return len(extensions[i]) > len(extensions[j])
		})

		overrides := make([]pathOverride, len(config.PathOverrides))
		for i, override := range config.PathOverrides {
			overrides[i] = newPathOverride(override)
		}

		return &ProjectLocator{
			modules:            make(map[string]struct{}),
			moduleCalls:        make(map[string][]string),
			discoveredVarFiles: make(map[string][]RootPathVarFile),
			excludedDirs:       config.ExcludedDirs,
			changedObjects:     config.ChangedObjects,
			includedDirs:       config.IncludedDirs,
			pathOverrides:      overrides,
			logger:             logger,
			envMatcher:         matcher,
			useAllPaths:        config.UseAllPaths,
			skip:               config.SkipAutoDetection,
			maxConfiguredDepth: config.MaxSearchDepth,
			shouldSkipDir: func(s string) bool {
				return false
			},
			shouldIncludeDir: func(s string) bool {
				return false
			},
			fallbackToIncludePaths:     config.FallbackToIncludePaths,
			forceProjectType:           config.ForceProjectType,
			terraformVarFileExtensions: extensions,
			hclParser:                  hclparse.NewParser(),
			hasCustomEnvExt:            len(config.TerraformVarFileExtensions) > 0,
			workingDirectory:           config.WorkingDirectory,
		}
	}

	return &ProjectLocator{
		modules:                    make(map[string]struct{}),
		discoveredVarFiles:         make(map[string][]RootPathVarFile),
		logger:                     logger,
		envMatcher:                 matcher,
		terraformVarFileExtensions: defaultExtensions,
		hclParser:                  hclparse.NewParser(),
	}
}

func (p *ProjectLocator) hasChanges(dir string) bool {
	if len(p.changedObjects) == 0 {
		return false
	}

	for _, change := range p.changedObjects {
		if inProject(dir, change) {
			return true
		}

		// let's check if any of the file changes are within this project's modules.
		for _, call := range p.moduleCalls[dir] {
			if inProject(call, change) {
				return true
			}
		}
	}

	return false
}

func inProject(dir string, change string) bool {
	rel, err := filepath.Rel(dir, change)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}

// TreeNode represents a node in the tree of Terraform projects. A TreeNode can
// either be a Terraform project, a directory containing Terraform var files, or
// just a filler node to represent a directory. Callers should check the RootPath
// and TerraformVarFiles fields to determine what type of node this is.
type TreeNode struct {
	Name              string
	Level             int
	RootPath          *RootPath
	TerraformVarFiles *VarFiles
	Children          []*TreeNode
	Parent            *TreeNode
}

// VarFiles represents a directory that contains Terraform var files. HasSiblings
// is true if the directory is within a directory that contains other directories
// that are root Terraform projects.
type VarFiles struct {
	Path  string
	Files []RootPathVarFile

	// HasSiblings is true if the directory is within a directory that contains other
	// root Terraform projects.
	HasSiblings bool
	// Used is true if the var files have been used by a project.
	Used bool
}

// CreateTreeNode creates a tree of Terraform projects and directories that
// contain var files.
func CreateTreeNode(basePath string, paths []RootPath, varFiles map[string][]RootPathVarFile, e *EnvFileMatcher) *TreeNode {
	root := &TreeNode{
		Name: "root",
	}

	sort.Slice(paths, func(i, j int) bool {
		return strings.Count(paths[i].DetectedPath, string(filepath.Separator)) < strings.Count(paths[j].DetectedPath, string(filepath.Separator))
	})

	for _, path := range paths {
		root.AddPath(path)
	}

	var varFilesSorted []string
	for dir := range varFiles {
		varFilesSorted = append(varFilesSorted, dir)
	}
	sort.Slice(varFilesSorted, func(i, j int) bool {
		return strings.Count(varFilesSorted[i], string(filepath.Separator)) < strings.Count(varFilesSorted[j], string(filepath.Separator))
	})

	for _, dir := range varFilesSorted {
		root.AddTerraformVarFiles(basePath, dir, varFiles[dir])
	}

	buildVarFileEnvNames(root, e)
	return root
}

// buildVarFileEnvNames builds the EnvName field for each var file. Var names
// can be inferred from the filename or from parent directories, but only if
// the parent directories don't contain any Terraform projects.
func buildVarFileEnvNames(root *TreeNode, e *EnvFileMatcher) {
	root.PostOrder(func(t *TreeNode) {
		if t.TerraformVarFiles == nil {
			return
		}

		parent := t.Parent
		current := t
		possibleEnvNames := []string{t.Name}
	c:
		for {
			if parent == nil {
				break
			}

			// only detect possible env names from directories that don't
			// have any Terraform projects in them. This means empty folders
			// or directories that wrap var file directories.
			for _, node := range parent.ChildNodesExcluding(current) {
				if node.RootPath != nil {
					break c
				}
			}

			possibleEnvNames = append(possibleEnvNames, parent.Name)
			current = parent
			parent = parent.Parent
		}

		for i, f := range t.TerraformVarFiles.Files {
			namesToSearch := append([]string{f.Name}, possibleEnvNames...)
			var envName string
			for _, search := range namesToSearch {
				if e.IsEnvName(search) {
					envName = search
					break
				}
			}

			t.TerraformVarFiles.Files[i].EnvName = e.EnvName(envName)
			if envName != "" {
				t.TerraformVarFiles.Files[i].IsGlobal = false
			}
		}
	})
}

// AddPath adds a path to the tree, this will create any missing nodes in the tree.
func (t *TreeNode) AddPath(path RootPath) {
	dir, _ := filepath.Rel(path.StartingPath, path.DetectedPath)

	pieces := strings.Split(dir, string(filepath.Separator))
	current := t
	for i, s := range pieces {
		if s == "" {
			continue
		}

		n := current.findChild(s)
		if n != nil {
			current = n
			continue
		}

		if i == len(pieces)-1 {
			break
		}

		child := &TreeNode{
			Name:   s,
			Level:  current.Level + 1,
			Parent: current,
		}

		current.Children = append(current.Children, child)
		current = child
	}

	current.Children = append(current.Children, &TreeNode{
		Name:     pieces[len(pieces)-1],
		Level:    current.Level + 1,
		RootPath: &path,
		Parent:   current,
	})
}

// findChild finds a child node with the given name.
func (t *TreeNode) findChild(name string) *TreeNode {
	for _, child := range t.Children {
		if child.Name == name {
			return child
		}
	}

	return nil
}

// ParentNode returns the parent node of the current node, this will skip any
// nodes that are not Terraform projects or directories that contain var files.
func (t *TreeNode) ParentNode() *TreeNode {
	if t.Parent != nil {
		if t.Parent.shouldVisitNode() {
			return t.Parent
		}

		if t.Parent.Name == "root" {
			return t.Parent
		}

		return t.Parent.ParentNode()
	}

	return nil
}

// UnusedParentVarFiles returns a list of any parent directories that contain var
// files that have not been used by a project.
func (t *TreeNode) UnusedParentVarFiles() []*VarFiles {
	parent := t.ParentNode()
	if parent == nil {
		return nil
	}

	var varFiles []*VarFiles
	if parent.TerraformVarFiles != nil && !parent.TerraformVarFiles.Used {
		varFiles = append(varFiles, parent.TerraformVarFiles)
	}

	return append(varFiles, parent.UnusedParentVarFiles()...)
}

// FindTfvarsCommonParent returns the first parent directory that has a child
// directory with a root Terraform project.
func (t *TreeNode) FindTfvarsCommonParent() *TreeNode {
	parent := t.Parent

	for {
		if parent == nil {
			return nil
		}

		for _, child := range parent.ChildNodesExcluding(t) {
			if child.RootPath != nil {
				return parent
			}
		}

		parent = parent.Parent
	}
}

// ChildNodesExcluding collects all the child nodes of the current node,
// excluding the given root node.
func (t *TreeNode) ChildNodesExcluding(root *TreeNode) []*TreeNode {
	var children []*TreeNode
	for _, child := range t.Children {
		if child.shouldVisitNode() && child != root {
			children = append(children, child)
		}
	}

	for _, child := range t.Children {
		if child != root {
			children = append(children, child.ChildNodesExcluding(root)...)
		}
	}

	return children
}

// ChildNodes returns the first set of child nodes that are Terraform projects or
// directories that contain var files.
func (t *TreeNode) ChildNodes() []*TreeNode {
	var children []*TreeNode
	for _, child := range t.Children {
		if child.shouldVisitNode() {
			children = append(children, child)
		}
	}

	if len(children) > 0 {
		return children
	}

	for _, child := range t.Children {
		children = append(children, child.ChildNodes()...)
	}

	return children
}

// ChildTerraformVarFiles returns the first set of child nodes that contain just terraform
// var files.
func (t *TreeNode) ChildTerraformVarFiles() []*TreeNode {
	var children []*TreeNode
	for _, child := range t.Children {
		if child.TerraformVarFiles != nil && child.RootPath == nil {
			children = append(children, child)
		}
	}

	if len(children) > 0 {
		return children
	}

	for _, child := range t.Children {
		children = append(children, child.ChildTerraformVarFiles()...)
	}

	return children
}

// AddTerraformVarFiles adds a directory that contains Terraform var files to the tree.
func (t *TreeNode) AddTerraformVarFiles(basePath, dir string, files []RootPathVarFile) {
	rel, _ := filepath.Rel(basePath, dir)
	pieces := strings.Split(rel, string(filepath.Separator))
	current := t
	for i, s := range pieces {
		if s == "" {
			continue
		}

		n := current.findChild(s)
		if n != nil {
			current = n
			continue
		}

		if i == len(pieces)-1 {
			break
		}

		child := &TreeNode{
			Name:   s,
			Level:  current.Level + 1,
			Parent: current,
		}

		current.Children = append(current.Children, child)
		current = child
	}

	var hasSiblings bool
	for _, child := range current.Children {
		if child.RootPath != nil && current.ParentNode() != nil {
			for _, node := range current.ParentNode().Children {
				if node.TerraformVarFiles != nil && (node.TerraformVarFiles.HasSiblings || current.ParentNode().Name == "root") {
					hasSiblings = true
					break
				}
			}
		}

		if hasSiblings {
			break
		}
	}

	if current.Name == pieces[len(pieces)-1] {
		current.TerraformVarFiles = &VarFiles{
			Path:        dir,
			Files:       files,
			HasSiblings: hasSiblings,
		}

		return
	}

	current.Children = append(current.Children, &TreeNode{
		Name: pieces[len(pieces)-1],
		TerraformVarFiles: &VarFiles{
			Path:        dir,
			Files:       files,
			HasSiblings: hasSiblings,
		},
		Parent: current,
		Level:  current.Level + 1,
	})
}

// PostOrder traverses the tree in post order, calling the given function on each
// node. This will skip any nodes that are not Terraform projects or directories
// that contain var files.
func (t *TreeNode) PostOrder(visit func(t *TreeNode)) {
	for _, child := range t.Children {
		child.PostOrder(visit)
	}

	if t.shouldVisitNode() {
		visit(t)
	}
}

// Visit traverses the tree in pre order, calling the given function on each
// node. This will skip any nodes that are not Terraform projects or directories
func (t *TreeNode) Visit(f func(t *TreeNode)) {
	f(t)

	for _, child := range t.Children {
		child.Visit(f)
	}
}

func (t *TreeNode) shouldVisitNode() bool {
	return t.RootPath != nil || t.TerraformVarFiles != nil
}

// AssociateChildVarFiles make sure that any projects have directories which
// contain var files are associated with the project. These are only associated
// if they are within 2 levels of the project and not if the child directory is a
// valid sibling directory.
func (t *TreeNode) AssociateChildVarFiles() {
	t.PostOrder(func(t *TreeNode) {
		if t.RootPath == nil {
			return
		}

		for _, child := range t.ChildTerraformVarFiles() {
			// if the child has already been associated with a project skip it as the var
			// directory has already been associated with a root module which is a closer
			// relation to it than the current root path.
			if child.TerraformVarFiles.Used {
				continue
			}

			depth, err := getChildDepth(t.RootPath.DetectedPath, child.TerraformVarFiles.Path)
			if depth > 2 || err != nil {
				continue
			}

			if child.TerraformVarFiles.HasSiblings {
				// visit all the children of this node and make sure that these siblings
				// don't have any children with var files in them.
				var hasChildVarFiles bool
				for _, treeNode := range t.ChildNodes() {
					if treeNode.RootPath != nil && treeNode.RootPath.HasChildVarFiles {
						hasChildVarFiles = true
						break
					}
				}

				if !hasChildVarFiles {
					return
				}
			}

			t.RootPath.HasChildVarFiles = true
			child.TerraformVarFiles.Used = true

			t.RootPath.AddVarFiles(child.TerraformVarFiles)
		}
	})
}

// AssociateSiblingVarFiles makes sure that any sibling directories that contain
// var files are associated with their corresponding projects.
func (t *TreeNode) AssociateSiblingVarFiles() {
	t.Visit(func(t *TreeNode) {
		var rootPaths []*TreeNode
		var varDirs []*TreeNode
		for _, node := range t.Children {
			if node.RootPath != nil {
				rootPaths = append(rootPaths, node)
			}

			if node.TerraformVarFiles != nil && !node.TerraformVarFiles.Used {
				varDirs = append(varDirs, node)
			}
		}

		for _, path := range rootPaths {
			if !path.RootPath.HasChildVarFiles {
				for _, dir := range varDirs {
					dir.TerraformVarFiles.Used = true

					path.RootPath.AddVarFiles(dir.TerraformVarFiles)
				}
			}
		}
	})
}

// AssociateParentVarFiles returns a list of any parent directories that contain var
// files that have not been used by a project.
func (t *TreeNode) AssociateParentVarFiles() {
	t.PostOrder(func(t *TreeNode) {
		if t.RootPath == nil {
			return
		}

		varFiles := t.UnusedParentVarFiles()
		for _, varFile := range varFiles {
			t.RootPath.AddVarFiles(varFile)
		}
	})
}

// AssociateAuntVarFiles returns a list of any aunt directories that contain var
// files that have not been used by a project.
func (t *TreeNode) AssociateAuntVarFiles() {
	t.PostOrder(func(t *TreeNode) {
		if t.RootPath == nil {
			return
		}

		varFiles := t.UnusedParentVarFiles()
		for _, varFile := range varFiles {
			varFile.Used = true
		}
	})

	t.PostOrder(func(t *TreeNode) {
		if t.TerraformVarFiles == nil || t.TerraformVarFiles.Used {
			return
		}

		commonParent := t.FindTfvarsCommonParent()
		if commonParent == nil {
			return
		}

		for _, node := range commonParent.ChildNodesExcluding(t) {
			if node.RootPath != nil {
				node.RootPath.AddVarFiles(t.TerraformVarFiles)
			}
		}

	})
}

// CollectRootPaths returns a list of all the Terraform projects found in the tree.
func (t *TreeNode) CollectRootPaths(e *EnvFileMatcher) []RootPath {
	var projects []RootPath
	t.Visit(func(t *TreeNode) {
		if t.RootPath != nil {
			projects = append(projects, *t.RootPath)
		}
	})

	for i := range projects {
		sort.Slice(projects[i].TerraformVarFiles, func(x, y int) bool {
			return projects[i].TerraformVarFiles[x].RelPath < projects[i].TerraformVarFiles[y].RelPath
		})
	}

	found := make(map[string]bool)
	for _, root := range projects {
		for _, varFile := range root.TerraformVarFiles {
			base := filepath.Base(root.DetectedPath)
			name := e.clean(varFile.Name)
			if base == name {
				found[varFile.FullPath] = true
			}
		}
	}

	// filter terraform var files from the root paths that have
	// the same name as another root directory. This means that
	// terraform var files that are scoped to a specific project
	// are not added to another project.
	for i, root := range projects {
		var filtered RootPathVarFiles
		for _, varFile := range root.TerraformVarFiles {
			name := e.clean(varFile.Name)
			base := filepath.Base(root.DetectedPath)
			if found[varFile.FullPath] && base != name {
				continue
			}

			filtered = append(filtered, varFile)
		}
		projects[i].TerraformVarFiles = filtered
	}

	return projects
}

// RootPath holds information about the root directory of a project, this is normally the top level
// Terraform containing provider blocks.
type RootPath struct {
	Matcher *EnvFileMatcher

	// StartingPath is the path to the directory where the search started.
	StartingPath string
	// DetectedPath is the path to the root of the project.
	DetectedPath string
	// HasChanges contains information about whether the project has git changes associated with it.
	// This will show as true if one or more files/directories have changed in the Path, and also if
	// and local modules that are used by this project have changes.
	HasChanges bool
	// TerraformVarFiles are a list of any .tfvars or .tfvars.json files found at the root level.
	TerraformVarFiles RootPathVarFiles

	HasChildVarFiles bool
	IsTerragrunt     bool
}

func (r *RootPath) RelPath() string {
	rel, _ := filepath.Rel(r.StartingPath, r.DetectedPath)
	return rel
}

// GlobalFiles returns a list of any global var files defined in the project.
func (r *RootPath) GlobalFiles() RootPathVarFiles {
	var files RootPathVarFiles

	for _, varFile := range r.TerraformVarFiles {
		if varFile.IsGlobal {
			files = append(files, varFile)
		}
	}

	return files
}

// AutoFiles returns a list of any auto.tfvars or terraform.tfvars files defined in the project.
func (r *RootPath) AutoFiles() RootPathVarFiles {
	var files RootPathVarFiles

	for _, varFile := range r.TerraformVarFiles {
		if IsAutoVarFile(varFile.EnvName) {
			files = append(files, varFile)
		}
	}

	return files
}

// EnvFiles returns a list of any environment specific var files defined in the project.
func (r *RootPath) EnvFiles() RootPathVarFiles {
	var files RootPathVarFiles

	for _, varFile := range r.TerraformVarFiles {
		if !IsAutoVarFile(varFile.EnvName) && !varFile.IsGlobal {
			files = append(files, varFile)
		}
	}

	return files
}

// VarFileGrouping defines a grouping of var files by environment.
type VarFileGrouping struct {
	Name              string
	TerraformVarFiles RootPathVarFiles
}

// EnvGroupings returns a list of var file groupings by environment. This is used
// to group and dedup var files that would otherwise create new projects.
func (r *RootPath) EnvGroupings() []VarFileGrouping {
	if r.Matcher == nil {
		r.Matcher = CreateEnvFileMatcher(defaultEnvs, nil)
	}

	varFiles := r.EnvFiles()
	varFileGrouping := map[string]RootPathVarFiles{}

	for _, varFile := range varFiles {
		// first add only terraform var files that are children of this project.
		if varFile.IsChildVarFile() {
			env := r.Matcher.EnvName(varFile.EnvName)
			varFileGrouping[env] = append(varFileGrouping[env], varFile)
		}
	}

	hasChildVarFileEnvs := len(varFileGrouping) > 0

	for _, varFile := range varFiles {
		if varFile.IsChildVarFile() {
			continue
		}

		env := r.Matcher.EnvName(varFile.EnvName)
		_, exists := varFileGrouping[env]
		// only add the non child env var files if there are no envs defined that are
		// closer to the project, or if the env matches one defined as a child var file.
		if !hasChildVarFileEnvs || (hasChildVarFileEnvs && exists) {
			varFileGrouping[env] = append(varFileGrouping[env], varFile)
		}
	}

	var envNames []string
	for env := range varFileGrouping {
		envNames = append(envNames, env)
	}
	sort.Strings(envNames)

	var varEnvs []VarFileGrouping
	for _, env := range envNames {
		varEnvs = append(varEnvs, VarFileGrouping{
			Name:              env,
			TerraformVarFiles: varFileGrouping[env],
		})
	}

	return varEnvs
}

type RootPathVarFile struct {
	Name string
	// RelPath is the path relative to the root of the project.
	RelPath string

	IsGlobal bool
	EnvName  string
	FullPath string
}

func (r RootPathVarFile) IsChildVarFile() bool {
	return !strings.HasPrefix(r.RelPath, "..")
}

type RootPathVarFiles []RootPathVarFile

func (r RootPathVarFiles) ToPaths() []string {
	var paths = make([]string, len(r))

	for i, varFile := range r {
		paths[i] = varFile.RelPath
	}

	return paths
}

func (r *RootPath) AddVarFiles(v *VarFiles) {
	rel, _ := filepath.Rel(r.DetectedPath, v.Path)

	for _, f := range v.Files {
		r.TerraformVarFiles = append(r.TerraformVarFiles, RootPathVarFile{
			FullPath: filepath.Join(v.Path, f.Name),
			Name:     f.Name,
			EnvName:  f.EnvName,
			RelPath:  filepath.Join(rel, f.Name),
			IsGlobal: f.IsGlobal,
		})
	}
}

// FindRootModules returns a list of all directories that contain a full
// Terraform project under the given fullPath. This list excludes any Terraform
// modules that have been found (if they have been called by a Module source).
func (p *ProjectLocator) FindRootModules(startingPath string) []RootPath {
	p.basePath, _ = filepath.Abs(startingPath)
	p.modules = make(map[string]struct{})
	p.projectDuplicates = make(map[string]bool)
	p.moduleCalls = make(map[string][]string)
	p.wdContainsTerragrunt = false
	p.discoveredProjects = []discoveredProject{}
	p.discoveredVarFiles = make(map[string][]RootPathVarFile)
	p.shouldSkipDir = buildDirMatcher(p.excludedDirs, startingPath)
	p.shouldIncludeDir = buildDirMatcher(p.includedDirs, startingPath)

	if p.skip {
		// if we are skipping auto-detection we just return the root path, but we still
		// want to walk the paths to find any auto.tfvars or terraform.tfvars files. So
		// let's just walk the top level directory.
		p.walkPaths(startingPath, 0, 1)
		p.findTerragruntDirs(startingPath)

		detectedPath := startingPath
		if p.workingDirectory != "" {
			startingPath = p.workingDirectory
		}

		return []RootPath{
			{
				StartingPath:      startingPath,
				DetectedPath:      detectedPath,
				IsTerragrunt:      p.wdContainsTerragrunt,
				TerraformVarFiles: p.discoveredVarFiles[startingPath],
			},
		}
	}

	p.findTerragruntDirs(startingPath)
	p.walkPaths(startingPath, 0, p.maxSearchDepth())
	for _, project := range p.discoveredProjects {
		if _, ok := p.projectDuplicates[project.path]; ok {
			p.projectDuplicates[project.path] = true
		} else {
			p.projectDuplicates[project.path] = false
		}
	}

	p.logger.Debug().Msgf("walking directory at %s returned a list of possible Terraform projects with length %d", startingPath, len(p.discoveredProjects))

	var projects []RootPath
	projectMap := map[string]bool{}
	for _, dir := range p.discoveredProjectsWithModulesFiltered() {
		if p.shouldUseProject(dir, false) {
			projects = append(projects, RootPath{
				StartingPath:      startingPath,
				DetectedPath:      dir.path,
				HasChanges:        p.hasChanges(dir.path),
				TerraformVarFiles: p.discoveredVarFiles[dir.path],
				Matcher:           p.envMatcher,
				IsTerragrunt:      dir.isTerragrunt,
			})
			projectMap[dir.path] = true
		}
	}

	if len(projects) == 0 && p.fallbackToIncludePaths {
		for _, dir := range p.discoveredProjectsWithModulesFiltered() {
			if p.shouldUseProject(dir, true) {
				projects = append(projects, RootPath{
					StartingPath:      startingPath,
					DetectedPath:      dir.path,
					HasChanges:        p.hasChanges(dir.path),
					TerraformVarFiles: p.discoveredVarFiles[dir.path],
					Matcher:           p.envMatcher,
					IsTerragrunt:      dir.isTerragrunt,
				})
				projectMap[dir.path] = true
			}
		}
	}

	for _, dir := range projects {
		delete(p.discoveredVarFiles, dir.DetectedPath)
	}

	node := CreateTreeNode(startingPath, projects, p.discoveredVarFiles, p.envMatcher)
	node.AssociateChildVarFiles()
	node.AssociateSiblingVarFiles()
	node.AssociateParentVarFiles()
	node.AssociateAuntVarFiles()

	paths := node.CollectRootPaths(p.envMatcher)
	for i := range paths {
		// if the locator has been configured with a working directory we need to change
		// the starting path of the root paths to the working directory. This means that
		// paths that have been defined in a config file are relative to the working
		// directory rather than the defined paths found in the config file.
		if p.workingDirectory != "" {
			paths[i].StartingPath = p.workingDirectory
		}
	}

	p.excludeEnvFromPaths(paths)

	sort.Slice(paths, func(i, j int) bool {
		return paths[i].DetectedPath < paths[j].DetectedPath
	})

	return paths
}

// excludeEnvFromPaths filters car files from the paths based on the path overrides.
func (p *ProjectLocator) excludeEnvFromPaths(paths []RootPath) {
	// filter the "only" paths first. This is done as "only" rules take precedence
	// over exclude rules. So if an env is defined in both only and exclude and
	// matches the same path, the "only" rule is the only one to apply.
	onlyPaths := map[string]struct{}{}
	for _, override := range p.pathOverrides {
		if len(override.only) > 0 {
			for i, path := range paths {
				relPath := path.RelPath()
				if override.glob.Match(relPath) {
					filtered := append(path.GlobalFiles(), path.AutoFiles()...)
					for _, varFile := range path.EnvFiles() {
						if _, ok := override.only[varFile.EnvName]; ok {
							onlyPaths[relPath+varFile.EnvName] = struct{}{}
							filtered = append(filtered, varFile)
						}
					}
					paths[i].TerraformVarFiles = filtered
				}
			}
		}
	}

	for _, override := range p.pathOverrides {
		if len(override.exclude) > 0 {
			for i, path := range paths {
				relPath := path.RelPath()
				if override.glob.Match(relPath) {
					var filtered RootPathVarFiles
					for _, varFile := range path.TerraformVarFiles {
						_, excluded := override.exclude[varFile.EnvName]
						_, only := onlyPaths[relPath+varFile.EnvName]
						if excluded && !only {
							continue
						}

						filtered = append(filtered, varFile)
					}

					paths[i].TerraformVarFiles = filtered
				}
			}
		}
	}
}

func (p *ProjectLocator) discoveredProjectsWithModulesFiltered() []discoveredProject {
	var projects []discoveredProject

	for _, dir := range p.discoveredProjects {
		if _, ok := p.modules[dir.path]; !ok || p.useAllPaths || p.shouldIncludeDir(dir.path) {
			projects = append(projects, dir)
		}
	}

	return projects
}

func (p *ProjectLocator) shouldUseProject(dir discoveredProject, force bool) bool {
	if p.shouldSkipDir(dir.path) {
		p.logger.Debug().Msgf("skipping directory %s as it is marked as excluded by --exclude-path", dir.path)

		return false
	}

	if force {
		return true
	}

	if p.shouldIncludeDir(dir.path) {
		return true
	}

	if p.shouldRemoveDuplicateProject(dir) {
		return false
	}

	// we only include Terraform projects that have been found alongside Terragrunt
	// projects if they have been forced to be included by --include-path. This is
	// done as we sometimes get collisions with the Terragrunt modules that are
	// incorrectly flagged as Terraform projects.
	//
	// @TODO in future we can read the "source" blocks of the Terragrunt projects and
	// infer that the Terraform projects are not modules.
	isForcedDupDir := p.projectDuplicates[dir.path] && p.forceProjectType != ""
	if !isForcedDupDir && p.wdContainsTerragrunt && !dir.isTerragrunt {
		return false
	}

	if p.useAllPaths {
		return true
	}

	if len(p.discoveredProjectsWithModulesFiltered()) == 1 {
		return true
	}

	if _, ok := p.modules[dir.path]; ok && !p.useAllPaths {
		p.logger.Debug().Msgf("skipping directory %s as it has been called as a module", dir.path)

		return false
	}

	if !dir.hasRootModuleBlocks() && !dir.isTerragrunt {
		return false
	}

	return true
}

// shouldRemoveDuplicateProject returns true if the project should be removed
// from the list of discovered projects based on the if the directory has been
// detected as both a terragrunt and terraform project. We only remove
// directories that have been flagged as duplicates and have been forced to be
// included as a specific project type.
func (p *ProjectLocator) shouldRemoveDuplicateProject(dir discoveredProject) bool {
	if p.forceProjectType == "" {
		return false
	}

	if !p.projectDuplicates[dir.path] {
		return false
	}

	if p.forceProjectType == "terragrunt" && !dir.isTerragrunt {
		return true
	}

	if p.forceProjectType == "terraform" && dir.isTerragrunt {
		return true
	}

	return false
}

// getChildDepth returns the number of levels targetPath is nested under
// basePath. If targetPath is not a child of basePath getChildDepth will return
// an error.
func getChildDepth(basePath string, targetPath string) (int, error) {
	rel, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		return 0, err
	}

	if strings.HasPrefix(rel, "..") || rel == "." {
		return 0, fmt.Errorf("%s is not a child of path %s", targetPath, basePath)
	}

	// Count the separators in the relative path
	return strings.Count(rel, string(filepath.Separator)) + 1, nil
}

func (p *ProjectLocator) maxSearchDepth() int {
	if p.maxConfiguredDepth > 0 {
		return p.maxConfiguredDepth
	}

	if p.useAllPaths {
		return 14
	}

	return 7
}

func (p *ProjectLocator) walkPaths(fullPath string, level int, maxSearchDepth int) {
	p.logger.Debug().Msgf("walking path %s to discover terraform files", fullPath)

	if level >= maxSearchDepth {
		p.logger.Debug().Msgf("exiting parsing directory %s as it is outside the maximum evaluation threshold", fullPath)
		return
	}

	hclParser := hclparse.NewParser()

	fileInfos, err := os.ReadDir(fullPath)
	if err != nil {
		p.logger.Debug().Err(err).Msgf("could not get file information for path %s skipping evaluation", fullPath)
		return
	}

	for _, info := range fileInfos {
		if info.IsDir() {
			continue
		}

		var parseFunc func(filename string) (*hcl.File, hcl.Diagnostics)
		name := info.Name()

		if strings.HasSuffix(name, ".tf") {
			parseFunc = hclParser.ParseHCLFile
		}

		if strings.HasSuffix(name, ".tf.json") {
			parseFunc = hclParser.ParseJSONFile
		}

		if p.isTerraformVarFile(name, filepath.Join(fullPath, name)) {
			v, ok := p.discoveredVarFiles[fullPath]
			if !ok {
				v = []RootPathVarFile{{
					Name:     name,
					EnvName:  p.envMatcher.EnvName(name),
					RelPath:  name,
					IsGlobal: p.envMatcher.IsGlobalVarFile(name),
					FullPath: filepath.Join(fullPath, name),
				}}
			} else {
				v = append(v, RootPathVarFile{
					Name:     name,
					EnvName:  p.envMatcher.EnvName(name),
					RelPath:  name,
					IsGlobal: p.envMatcher.IsGlobalVarFile(name),
					FullPath: filepath.Join(fullPath, name),
				})
			}

			p.discoveredVarFiles[fullPath] = v
		}

		if parseFunc == nil {
			continue
		}

		path := filepath.Join(fullPath, name)
		_, diag := parseFunc(path)
		if diag != nil && diag.HasErrors() {
			p.logger.Debug().Msgf("skipping file: %s hcl parsing err: %s", path, diag.Error())
			continue
		}
	}

	files := hclParser.Files()
	if len(files) > 0 {
		blockInfo := p.shallowDecodeTerraformBlocks(fullPath, files)

		p.discoveredProjects = append(p.discoveredProjects, discoveredProject{
			path:             fullPath,
			files:            files,
			hasProviderBlock: blockInfo.hasProviderBlock,
			hasBackendBlock:  blockInfo.hasTerraformBackendBlock,
			depth:            level,
		})
	}

	for _, info := range fileInfos {
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") {
				continue
			}

			p.walkPaths(filepath.Join(fullPath, info.Name()), level+1, maxSearchDepth)
		}
	}
}

func (p *ProjectLocator) isTerraformVarFile(name string, fullPath string) bool {
	if hasDefaultVarFileExtension(name) {
		return true
	}

	// we also check for tfvars.json files as these are non-standard naming
	// conventions which are used by some projects.
	if strings.HasPrefix(name, "tfvars") && strings.HasSuffix(name, ".json") {
		return true
	}

	// If there are no custom extensions we can early exit here.
	if !p.hasCustomEnvExt {
		return false
	}

	// if we have custom extensions enabled in the autodetect configuration we need
	// to check the extension of the file to see if it matches any of the custom
	if !hasVarFileExtension(name, p.terraformVarFileExtensions) {
		return false
	}

	// if we have custom extensions enabled in the autodetect configuration we need
	// to make sure that this file is a valid HCL file before we add it to the list
	// of discovered var files. This is because we can have collisions with custom
	// env var extensions and other files that are not valid HCL files. e.g. with an
	// empty/wildcard extension we could match a file called "tfvars" and also
	// "Jenkinsfile", the latter being a non-HCL file.
	f, d := p.hclParser.ParseHCLFile(fullPath)
	if d != nil {
		return false
	}

	// If the file is empty or has a comment, it would still be considered a valid
	// So we check it has at least one attribute defined.
	attr, _ := f.Body.JustAttributes()

	return len(attr) > 0
}

func hasDefaultVarFileExtension(name string) bool {
	for _, extension := range defaultExtensions {
		if strings.HasSuffix(name, extension) {
			return true
		}
	}

	return false
}

// splitVarFileExt splits the var file extension (.tfvar, .tfvar.json, etc) from a file name.
// It will return the file name without the extension and the extension itself. If the file name
// does not have a valid var file extension it will return the original file name and an empty string.
//
// The valid extensions should be passed in by the caller sorted by preference.
func splitVarFileExt(fileName string, sortedExts []string) (string, string) {
	if len(fileName) == 0 {
		return "", ""
	}

	for _, ext := range sortedExts {
		if strings.HasSuffix(fileName, ext) {
			return fileName[:len(fileName)-len(ext)], ext
		}
	}

	return fileName, ""
}

// hasVarFileExtension checks if the file name has a valid extension.
func hasVarFileExtension(fileName string, extensions []string) bool {
	if len(fileName) == 0 {
		return false
	}

	_, varFileExt := splitVarFileExt(fileName, extensions)
	if varFileExt != "" {
		return true
	}

	// Check if a "" extension is allowed. This means we allowed var files to be in files
	// without an extension such as `prod` or `dev`.
	blankExtensionAllowed := false
	for _, e := range extensions {
		if e == "" {
			blankExtensionAllowed = true
			break
		}
	}

	// If the file has no extension and we allow blank extensions we return true.
	// When checking if the extension is blank we also remove any leading dots from
	// the file name otherwise filepath.Ext returns the full file name as the extension
	// for hidden files.
	if filepath.Ext(strings.TrimPrefix(fileName, ".")) == "" && blankExtensionAllowed {
		return true
	}

	return false
}

type terraformDirInfo struct {
	hasProviderBlock         bool
	hasTerraformBackendBlock bool
}

func (p *ProjectLocator) shallowDecodeTerraformBlocks(fullPath string, files map[string]*hcl.File) terraformDirInfo {
	var hasProviderBlock bool
	var hasTerraformBackendBlock bool

	for _, file := range files {
		// file can potential be nil if it could not be opened or read when loading.
		// This can happen if the user runs against a Terraform project with invalid JSON files.
		if file == nil {
			continue
		}

		body, content, diags := file.Body.PartialContent(terraformAndProviderBlocks)
		if diags != nil && diags.HasErrors() {
			p.logger.Debug().Err(diags).Msgf("skipping building module information for file %s as failed to get partial body contents", file)
			continue
		}

		providerBlocks := body.Blocks.OfType("provider")
		if len(providerBlocks) > 0 {
			hasProviderBlock = true
		}

		terraformBlocks := body.Blocks.OfType("terraform")
		for _, block := range terraformBlocks {
			backend, _, _ := block.Body.PartialContent(nestedBackendBlock)
			if len(backend.Blocks) > 0 {
				hasTerraformBackendBlock = true
				break
			}
		}

		moduleBody, _, _ := content.PartialContent(justModuleBlocks)
		for _, module := range moduleBody.Blocks {
			a, _ := module.Body.JustAttributes()
			if src, ok := a["source"]; ok {
				val, _ := src.Expr.Value(nil)

				if val.Type() != cty.String {
					p.logger.Debug().Str("module", strings.Join(module.Labels, ".")).Msgf("got unexpected cty value for module source string in file %s", file)
					continue
				}

				var realPath string
				err := gocty.FromCtyValue(val, &realPath)
				if err != nil {
					p.logger.Debug().Err(err).Str("module", strings.Join(module.Labels, ".")).Msg("could not read source value of module as string")
					continue
				}

				mp := filepath.Join(fullPath, realPath)
				p.modules[mp] = struct{}{}
				if v, ok := p.moduleCalls[fullPath]; ok {
					p.moduleCalls[fullPath] = append(v, mp)
				} else {
					p.moduleCalls[fullPath] = []string{mp}
				}
			}
		}
	}
	return terraformDirInfo{
		hasProviderBlock:         hasProviderBlock,
		hasTerraformBackendBlock: hasTerraformBackendBlock,
	}
}

func (p *ProjectLocator) findTerragruntDirs(fullPath string) {
	terragruntCacheDir := filepath.Join(config.InfracostDir, ".terragrunt-cache")
	terragruntDownloadDir := filepath.Join(fullPath, terragruntCacheDir)
	terragruntConfigFiles, err := tgconfig.FindConfigFilesInPath(fullPath, &options.TerragruntOptions{
		DownloadDir: terragruntDownloadDir,
	})
	if err != nil {
		p.logger.Debug().Err(err).Msgf("failed to find terragrunt files in path %s", fullPath)
	}

	if len(terragruntConfigFiles) > 0 {
		p.wdContainsTerragrunt = true
	}

	for _, configFile := range terragruntConfigFiles {
		if !p.shouldSkipDir(filepath.Dir(configFile)) && !IsParentTerragruntConfig(configFile, terragruntConfigFiles) {
			p.discoveredProjects = append(p.discoveredProjects, discoveredProject{
				path:         filepath.Dir(configFile),
				isTerragrunt: true,
			})
		}
	}
}

func buildDirMatcher(dirs []string, fullPath string) func(string) bool {
	var rawMatches []string
	globMatches := make(map[string]struct{})

	for _, dir := range dirs {
		var absoluteDir string
		if dir == filepath.Base(dir) {
			rawMatches = append(rawMatches, dir)
		}

		if filepath.IsAbs(dir) {
			absoluteDir = dir
		} else {
			absoluteDir = filepath.Join(fullPath, dir)
		}

		globs, err := filepath.Glob(absoluteDir)
		if err == nil {
			for _, m := range globs {
				globMatches[m] = struct{}{}
			}
		}
	}

	return func(dir string) bool {
		if _, ok := globMatches[dir]; ok {
			return true
		}

		base := filepath.Base(dir)
		for _, match := range rawMatches {
			if match == base {
				return true
			}
		}

		return false
	}
}

// IsParentTerragruntConfig checks if a terragrunt config entry is a parent file that is referenced by another config
// with a find_in_parent_folders call. The find_in_parent_folders function searches up the directory tree
// from the file and returns the absolute path to the first terragrunt.hcl. This means if it is found
// we can treat this file as a child terragrunt.hcl.
func IsParentTerragruntConfig(parent string, configFiles []string) bool {
	for _, name := range configFiles {
		if !isChildDirectory(parent, name) {
			continue
		}

		file, err := os.Open(name)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			// skip any commented out lines
			if strings.HasPrefix(line, "#") {
				continue
			}

			if strings.Contains(line, "find_in_parent_folders()") {
				file.Close()
				return true
			}
		}

		file.Close()
	}

	return false
}

func isChildDirectory(parent, child string) bool {
	if parent == child {
		return false
	}

	parentDir := filepath.Dir(parent)
	childDir := filepath.Dir(child)
	p, err := filepath.Rel(parentDir, childDir)
	if err != nil || strings.Contains(p, "..") {
		return false
	}

	return true
}
