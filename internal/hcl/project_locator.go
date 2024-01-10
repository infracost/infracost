package hcl

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/rs/zerolog"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

var (
	defaultEnvVarNames = regexp.MustCompile(`^(prd|prod|production|preprod|staging|stage|stg|development|dev|release|testing|test|tst|qa|uat|live|sandbox|demo|integration|int|experimental|experiments|trial|validation|perf|sec|dr)`)

	VarFileEnvPrefixRegxp = regexp.MustCompile(`^(\w+)-`)
)

// CleanVarName removes the .tfvars or .tfvars.json suffix from the file name.
func CleanVarName(file string) string {
	return filepath.Base(strings.TrimSuffix(strings.TrimSuffix(file, ".json"), ".tfvars"))
}

// VarEnvName returns the environment prefix of the clean name var file, if it
// has one.
func VarEnvName(file string) string {
	name := CleanVarName(file)
	sub := VarFileEnvPrefixRegxp.FindStringSubmatch(name)
	if len(sub) != 2 {
		return name
	}

	return sub[1]
}

// IsAutoVarFile checks if the var file is an auto.tfvars or terraform.tfvars.
// These are special Terraform var files that are applied to every project
// automatically.
func IsAutoVarFile(file string) bool {
	withoutJSONSuffix := strings.TrimSuffix(file, ".json")

	return strings.HasSuffix(withoutJSONSuffix, ".auto.tfvars") || withoutJSONSuffix == "terraform.tfvars"
}

type discoveredProject struct {
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
	modules        map[string]struct{}
	moduleCalls    map[string][]string
	excludedDirs   []string
	changedObjects []string
	useAllPaths    bool
	logger         zerolog.Logger

	basePath           string
	discoveredVarFiles map[string][]string
	discoveredProjects []discoveredProject
	includedDirs       []string
	envNames           *regexp.Regexp
	skip               bool
}

// ProjectLocatorConfig provides configuration options on how the locator functions.
type ProjectLocatorConfig struct {
	ExcludedDirs      []string
	ChangedObjects    []string
	UseAllPaths       bool
	SkipAutoDetection bool
	IncludedDirs      []string
	EnvNames          []string
}

// NewProjectLocator returns safely initialized ProjectLocator.
func NewProjectLocator(logger zerolog.Logger, config *ProjectLocatorConfig) *ProjectLocator {
	envVarNames := defaultEnvVarNames
	if config != nil {
		if len(config.EnvNames) > 0 {
			envVarNames = regexp.MustCompile("^" + strings.Join(config.EnvNames, "|"))
		}

		return &ProjectLocator{
			modules:            make(map[string]struct{}),
			moduleCalls:        make(map[string][]string),
			discoveredVarFiles: make(map[string][]string),
			excludedDirs:       config.ExcludedDirs,
			changedObjects:     config.ChangedObjects,
			includedDirs:       config.IncludedDirs,
			logger:             logger,
			envNames:           envVarNames,
			useAllPaths:        config.UseAllPaths,
			skip:               config.SkipAutoDetection,
		}
	}

	return &ProjectLocator{
		modules:            make(map[string]struct{}),
		discoveredVarFiles: make(map[string][]string),
		logger:             logger,
		envNames:           envVarNames,
	}
}

// IsGlobalVarFile checks if the var file is a "global" one, this is only
// applicable if we match a globalTerraformVarName and these don't have an
// environment prefix e.g. defaults.tfvars, global.tfvars are applicable,
// prod-default.tfvars, stag-globals are not.
func (p *ProjectLocator) IsGlobalVarFile(file string) bool {
	return !p.envNames.MatchString(filepath.Base(file))
}

func (p *ProjectLocator) buildSkippedMatcher(fullPath string) func(string) bool {
	var excludedMatches []string
	excludedGlobs := make(map[string]struct{})

	for _, dir := range p.excludedDirs {
		var absoluteDir string
		if dir == filepath.Base(dir) {
			excludedMatches = append(excludedMatches, dir)
		}

		if filepath.IsAbs(dir) {
			absoluteDir = dir
		} else {
			absoluteDir = filepath.Join(fullPath, dir)
		}

		globs, err := filepath.Glob(absoluteDir)
		if err == nil {
			for _, m := range globs {
				excludedGlobs[m] = struct{}{}
			}
		}
	}

	return func(dir string) bool {
		if _, ok := excludedGlobs[dir]; ok {
			return true
		}

		base := filepath.Base(dir)
		for _, match := range excludedMatches {
			if match == base {
				return true
			}
		}

		return false
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
	Files []string

	// HasSiblings is true if the directory is within a directory that contains other
	// root Terraform projects.
	HasSiblings bool
	// Used is true if the var files have been used by a project.
	Used bool
}

// CreateTreeNode creates a tree of Terraform projects and directories that
// contain var files.
func CreateTreeNode(basePath string, paths []RootPath, varFiles map[string][]string) *TreeNode {
	root := &TreeNode{
		Name: "root",
	}

	sort.Slice(paths, func(i, j int) bool {
		return strings.Count(paths[i].Path, string(filepath.Separator)) < strings.Count(paths[j].Path, string(filepath.Separator))
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

	return root
}

// AddPath adds a path to the tree, this will create any missing nodes in the tree.
func (t *TreeNode) AddPath(path RootPath) {
	dir, _ := filepath.Rel(path.RepoPath, path.Path)

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

// UnusedAuntVarFiles returns a list of any aunt directories that contain var
// files that have not been used by a project.
func (t *TreeNode) UnusedAuntVarFiles() []*VarFiles {
	// if we don't have a root path, we can't have an aunt.
	if t.RootPath == nil {
		return nil
	}

	parent := t.Parent
	if parent == nil {
		return nil
	}

	// get the parent of the parent as this will be above the current node.
	parent = t.Parent
	var varFiles []*VarFiles

	for {
		if parent == nil {
			return varFiles
		}

		for _, child := range parent.Children {
			if child.TerraformVarFiles != nil && !child.TerraformVarFiles.Used && child.RootPath == nil {
				varFiles = append(varFiles, child.TerraformVarFiles)
			}
		}

		parent = parent.Parent
	}
}

// ChildNodes returns a list of child nodes that are Terraform projects or
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

// AddTerraformVarFiles adds a directory that contains Terraform var files to the tree.
func (t *TreeNode) AddTerraformVarFiles(basePath, dir string, files []string) {
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

		for _, child := range t.ChildNodes() {
			if child.TerraformVarFiles == nil {
				continue
			}

			depth, err := getChildDepth(t.RootPath.Path, child.TerraformVarFiles.Path)
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

			t.RootPath.AddVarFiles(child.TerraformVarFiles.Path, child.TerraformVarFiles.Files)
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

					path.RootPath.AddVarFiles(dir.TerraformVarFiles.Path, dir.TerraformVarFiles.Files)
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
			t.RootPath.AddVarFiles(varFile.Path, varFile.Files)
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
		if t.RootPath == nil {
			return
		}

		varFiles := t.UnusedAuntVarFiles()
		for _, varFile := range varFiles {
			t.RootPath.AddVarFiles(varFile.Path, varFile.Files)
		}
	})
}

// CollectRootPaths returns a list of all the Terraform projects found in the tree.
func (t *TreeNode) CollectRootPaths() []RootPath {
	var projects []RootPath
	t.Visit(func(t *TreeNode) {
		if t.RootPath != nil {
			projects = append(projects, *t.RootPath)
		}
	})

	for i := range projects {
		sort.Strings(projects[i].TerraformVarFiles)
	}

	return projects
}

// RootPath holds information about the root directory of a project, this is normally the top level
// Terraform containing provider blocks.
type RootPath struct {
	RepoPath string
	Path     string
	// HasChanges contains information about whether the project has git changes associated with it.
	// This will show as true if one or more files/directories have changed in the Path, and also if
	// and local modules that are used by this project have changes.
	HasChanges bool
	// TerraformVarFiles are a list of any .tfvars or .tfvars.json files found at the root level.
	TerraformVarFiles TerraformVarFiles

	HasChildVarFiles bool
}

func (r *RootPath) AddVarFiles(dir string, files []string) {
	rel, _ := filepath.Rel(r.Path, dir)

	for _, f := range files {
		r.TerraformVarFiles.Add(filepath.Join(rel, f))
	}
}

type TerraformVarFiles []string

func (vf *TerraformVarFiles) Add(file string) {
	for _, f := range *vf {
		if f == file {
			return
		}
	}

	*vf = append(*vf, file)
}

// FindRootModules returns a list of all directories that contain a full Terraform project under the given fullPath.
// This list excludes any Terraform modules that have been found (if they have been called by a Module source).
func (p *ProjectLocator) FindRootModules(fullPath string) []RootPath {
	if p.skip {
		return []RootPath{
			{
				Path: fullPath,
			},
		}
	}

	p.basePath, _ = filepath.Abs(fullPath)
	p.modules = make(map[string]struct{})
	p.moduleCalls = make(map[string][]string)

	isSkipped := p.buildSkippedMatcher(fullPath)
	p.walkPaths(fullPath, 0)
	p.logger.Debug().Msgf("walking directory at %s returned a list of possible Terraform projects with length %d", fullPath, len(p.discoveredProjects))

	var projects []RootPath
	projectMap := map[string]bool{}
	for _, dir := range p.discoveredProjectsWithModulesFiltered() {
		if p.shouldUseProject(dir, isSkipped) {
			projects = append(projects, RootPath{
				RepoPath:          fullPath,
				Path:              dir.path,
				HasChanges:        p.hasChanges(dir.path),
				TerraformVarFiles: p.discoveredVarFiles[dir.path],
			})
			projectMap[dir.path] = true

			delete(p.discoveredVarFiles, dir.path)
		}
	}

	// add the user flagged included directories to the list of projects.
	for _, dir := range p.includedDirs {
		abs := path.Join(fullPath, dir)
		if _, err := os.Stat(abs); err != nil {
			continue
		}

		if !projectMap[abs] {
			projects = append(projects, RootPath{
				RepoPath: fullPath,
				Path:     abs,
			})
			projectMap[abs] = true
		}
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Path < projects[j].Path
	})

	node := CreateTreeNode(fullPath, projects, p.discoveredVarFiles)
	node.AssociateChildVarFiles()
	node.AssociateSiblingVarFiles()
	node.AssociateParentVarFiles()
	node.AssociateAuntVarFiles()

	return node.CollectRootPaths()
}

func (p *ProjectLocator) discoveredProjectsWithModulesFiltered() []discoveredProject {
	var projects []discoveredProject

	for _, dir := range p.discoveredProjects {
		if _, ok := p.modules[dir.path]; !ok || p.useAllPaths {
			projects = append(projects, dir)
		}
	}

	return projects

}

func (p *ProjectLocator) shouldUseProject(dir discoveredProject, isSkipped func(string) bool) bool {
	if isSkipped(dir.path) {
		p.logger.Debug().Msgf("skipping directory %s as it is marked as excluded by --exclude-path", dir.path)

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

	if !dir.hasRootModuleBlocks() {
		return false
	}

	return true
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
	if p.useAllPaths {
		return 14
	}

	return 7
}

func (p *ProjectLocator) walkPaths(fullPath string, level int) {
	// if the level is 0 this is the start of the directory tree.
	// let's reset all the discovered paths, so we don't duplicate.
	if level == 0 {
		p.discoveredProjects = []discoveredProject{}
		p.discoveredVarFiles = make(map[string][]string)
	}
	p.logger.Debug().Msgf("walking path %s to discover terraform files", fullPath)

	if level >= p.maxSearchDepth() {
		p.logger.Debug().Msgf("exiting parsing directory %s as it is outside the maximum evaluation threshold", fullPath)
		return
	}

	hclParser := hclparse.NewParser()

	fileInfos, err := os.ReadDir(fullPath)
	if err != nil {
		p.logger.Warn().Err(err).Msgf("could not get file information for path %s skipping evaluation", fullPath)
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

		if strings.HasSuffix(name, ".tfvars") || strings.HasSuffix(name, ".tfvars.json") {
			v, ok := p.discoveredVarFiles[fullPath]
			if !ok {
				v = []string{name}
			} else {
				v = append(v, name)
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
	var hasProviderBlock bool
	var hasTerraformBackendBlock bool

	for _, file := range files {
		body, content, diags := file.Body.PartialContent(terraformAndProviderBlocks)
		if diags != nil && diags.HasErrors() {
			p.logger.Warn().Err(diags).Msgf("skipping building module information for file %s as failed to get partial body contents", file)
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

	if len(files) > 0 {
		p.discoveredProjects = append(p.discoveredProjects, discoveredProject{
			path:             fullPath,
			files:            files,
			hasProviderBlock: hasProviderBlock,
			hasBackendBlock:  hasTerraformBackendBlock,
			depth:            level,
		})
	}

	for _, info := range fileInfos {
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") {
				continue
			}

			p.walkPaths(filepath.Join(fullPath, info.Name()), level+1)
		}
	}
}
