package hcl

import (
	"fmt"
	"os"
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
	// GlobalTerraformVarFileNames is a list of var file naming convention that suggests they are applied
	// to every project, despite changes in environment.
	GlobalTerraformVarFileNames = []string{
		"default",
		"defaults",
		"global",
		"globals",
		"shared",
	}

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

// IsGlobalVarFile checks if the var file is a "global" one, this is only
// applicable if we match a globalTerraformVarName and these don't have an
// environment prefix e.g. defaults.tfvars, global.tfvars are applicable,
// prod-default.tfvars, stag-globals are not.
func IsGlobalVarFile(file string) bool {
	name := CleanVarName(file)

	for _, global := range GlobalTerraformVarFileNames {
		if strings.HasSuffix(name, global) && !VarFileEnvPrefixRegxp.MatchString(name) {
			return true
		}
	}

	return false
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
}

// ProjectLocatorConfig provides configuration options on how the locator functions.
type ProjectLocatorConfig struct {
	ExcludedSubDirs   []string
	ChangedObjects    []string
	UseAllPaths       bool
	SkipAutoDetection bool
}

// NewProjectLocator returns safely initialized ProjectLocator.
func NewProjectLocator(logger zerolog.Logger, config *ProjectLocatorConfig) *ProjectLocator {
	if config != nil {
		return &ProjectLocator{
			modules:            make(map[string]struct{}),
			moduleCalls:        make(map[string][]string),
			discoveredVarFiles: make(map[string][]string),
			// by default we always exclude the "examples" directory as these are often found in
			// remote modules and can be valid projects, which are not used.
			excludedDirs:   append(config.ExcludedSubDirs, "examples"),
			changedObjects: config.ChangedObjects,
			logger:         logger,
			useAllPaths:    config.UseAllPaths,
		}
	}

	return &ProjectLocator{
		modules:            make(map[string]struct{}),
		discoveredVarFiles: make(map[string][]string),
		logger:             logger,
	}
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

func (vf *TerraformVarFiles) HasEnvFiles() bool {
	return len(*vf) > len(vf.GlobalFiles())
}

func (vf *TerraformVarFiles) GlobalFiles() TerraformVarFiles {
	var globals TerraformVarFiles

	for _, file := range *vf {
		if IsGlobalVarFile(file) {
			globals = append(globals, file)
		}
	}

	return globals
}

// FindRootModules returns a list of all directories that contain a full Terraform project under the given fullPath.
// This list excludes any Terraform modules that have been found (if they have been called by a Module source).
func (p *ProjectLocator) FindRootModules(fullPath string) []RootPath {
	p.basePath, _ = filepath.Abs(fullPath)
	p.modules = make(map[string]struct{})
	p.moduleCalls = make(map[string][]string)

	isSkipped := p.buildSkippedMatcher(fullPath)
	p.walkPaths(fullPath, 0)
	p.logger.Debug().Msgf("walking directory at %s returned a list of possible Terraform projects with length %d", fullPath, len(p.discoveredProjects))

	var projects []RootPath
	for _, dir := range p.discoveredProjects {
		if p.shouldUseProject(dir, isSkipped) {
			projects = append(projects, RootPath{
				RepoPath:          fullPath,
				Path:              dir.path,
				HasChanges:        p.hasChanges(dir.path),
				TerraformVarFiles: p.discoveredVarFiles[dir.path],
			})

			delete(p.discoveredVarFiles, dir.path)
		}
	}

	// loop through the remaining discovered var files that aren't at the same
	// directory as an existing project.

	// if the directory has var files in use that and exclude the directory
	// if the directory has sibling folders with var files use that
	// if the directory has child folders with var files use that
	// if the directory has parent folders with var files use that

	// loop through the remaining discovered var files that aren't at the same
	// directory as an existing project. If the directory is a sibling of an existing
	// project, but has not already been used as a child var file directory then we
	// associate the var files with the project.
	excludeVarDirectory := map[string]bool{}
	var varFiles []string
	for dir, _ := range p.discoveredVarFiles {
		varFiles = append(varFiles, dir)
	}

	for dir, files := range p.filteredVarFiles(excludeVarDirectory) {
		for i, project := range projects {
			if isSiblingDirRec(project.Path, dir, projects, varFiles) {
				p.logger.Debug().Msgf("found sibling directory %s to project %s", dir, project.Path)

				projects[i].AddVarFiles(dir, files)

				excludeVarDirectory[dir] = true
			}
		}
	}

	for dir, files := range p.filteredVarFiles(excludeVarDirectory) {
		for i, project := range projects {
			if isNestedDir(project.Path, dir, projects, 2, varFiles) {
				p.logger.Debug().Msgf("found child directory %s to project %s", dir, project.Path)

				projects[i].AddVarFiles(dir, files)

				excludeVarDirectory[dir] = true
			}
		}
	}

	// parent directories
	for dir, files := range p.filteredVarFiles(excludeVarDirectory) {
		for i, project := range projects {
			if isParentDir(dir, project.Path) {
				p.logger.Debug().Msgf("found parent directory %s to project %s", dir, project.Path)

				projects[i].AddVarFiles(dir, files)

				excludeVarDirectory[dir] = true
			}
		}
	}

	// aunt directories
	for dir, files := range p.filteredVarFiles(excludeVarDirectory) {
		for i, project := range projects {
			if isParentDir(filepath.Dir(dir), project.Path) {
				p.logger.Debug().Msgf("found aunt directory %s to project %s", dir, project.Path)

				projects[i].AddVarFiles(dir, files)
			}
		}
	}

	for i := range projects {
		sort.Strings(projects[i].TerraformVarFiles)
	}

	return projects
}

func (p *ProjectLocator) filteredVarFiles(excludeVarDirectory map[string]bool) map[string][]string {
	varFileDirs := map[string][]string{}

	for dir, files := range p.discoveredVarFiles {
		if !excludeVarDirectory[dir] {
			varFileDirs[dir] = files
		}
	}

	return varFileDirs
}

func isParentDir(parentDir, childDir string) bool {
	absParentDir, err := filepath.Abs(parentDir)
	if err != nil {
		return false
	}

	absChildDir, err := filepath.Abs(childDir)
	if err != nil {
		return false
	}

	if absChildDir == absParentDir {
		return false
	}

	return strings.HasPrefix(absChildDir, absParentDir)
}

func isSiblingDir(dir1 string, dir2 string) bool {
	return filepath.Dir(dir1) == filepath.Dir(dir2)
}

func isSiblingDirRec(projectDir string, varDir string, projects []RootPath, varDirs []string) bool {
	if !isSiblingDir(projectDir, varDir) {
		return false
	}

	parent := filepath.Dir(projectDir)
	for _, dir := range varDirs {
		if !isSiblingDirRec(parent, dir, projects, varDirs) {
			return false
		}
	}

	var filteredVarDirs []string
	for _, d := range varDirs {
		if varDir != d {
			filteredVarDirs = append(filteredVarDirs, d)
		}
	}

	for _, project := range projects {
		if !isSiblingDir(project.Path, varDir) {
			continue
		}

		for _, d := range filteredVarDirs {
			if isNestedDir(project.Path, d, projects, 1, varDirs) {
				return false
			}
		}
	}

	return true
}

func (p *ProjectLocator) shouldUseProject(dir discoveredProject, isSkipped func(string) bool) bool {
	if isSkipped(dir.path) {
		p.logger.Debug().Msgf("skipping directory %s as it is marked as excluded by --exclude-path", dir)

		return false
	}

	if p.useAllPaths {
		return true
	}

	if len(p.discoveredProjects) == 1 {
		return true
	}

	if _, ok := p.modules[dir.path]; ok && !p.useAllPaths {
		p.logger.Debug().Msgf("skipping directory %s as it has been called as a module", dir)

		return false
	}

	if !dir.hasRootModuleBlocks() {
		return false
	}

	return true
}

// isNestedDir checks if the target path nested no more than 'levels' under the
// base path and that no other project paths are a closer parent to the
// targetPath.
func isNestedDir(basePath, targetPath string, projects []RootPath, levels int, varFiles []string) bool {
	sepCount, err := getChildDepth(basePath, targetPath)
	if err != nil {
		return false
	}

	if sepCount > levels {
		return false
	}

	for _, project := range projects {
		if basePath == project.Path {
			continue
		}

		depth, err := getChildDepth(project.Path, targetPath)
		if err != nil {
			continue
		}

		if depth < sepCount {
			return false
		}
	}

	var filteredProjects []RootPath
	for _, project := range projects {
		if project.Path != basePath {
			filteredProjects = append(filteredProjects, project)
		}
	}

	for _, project := range filteredProjects {
		if isSiblingDirRec(project.Path, targetPath, filteredProjects, varFiles) {
			return false
		}
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
