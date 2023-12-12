package hcl

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/rs/zerolog"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

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
	discoveredProjects []string
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
	TerraformVarFiles []string
}

// FindRootModules returns a list of all directories that contain a full Terraform project under the given fullPath.
// This list excludes any Terraform modules that have been found (if they have been called by a Module source).
func (p *ProjectLocator) FindRootModules(fullPath string) []RootPath {
	p.basePath, _ = filepath.Abs(fullPath)
	p.modules = make(map[string]struct{})
	p.moduleCalls = make(map[string][]string)

	isSkipped := p.buildSkippedMatcher(fullPath)
	p.walkPaths(fullPath, 0)
	p.logger.Debug().Msgf("walking directory at %s returned a list of possible Terraform projects: %+v", fullPath, p.discoveredProjects)

	var projects []RootPath
	for _, dir := range p.discoveredProjects {
		if isSkipped(dir) {
			p.logger.Debug().Msgf("skipping directory %s as it is marked as excluded by --exclude-path", dir)
			continue
		}

		if _, ok := p.modules[dir]; ok && !p.useAllPaths {
			p.logger.Debug().Msgf("skipping directory %s as it has been called as a module", dir)
			continue
		}

		projects = append(projects, RootPath{
			RepoPath:          fullPath,
			Path:              dir,
			HasChanges:        p.hasChanges(dir),
			TerraformVarFiles: p.discoveredVarFiles[dir],
		})
		delete(p.discoveredVarFiles, dir)
	}

	// loop through the remaining discovered var files that aren't at the same
	// directory as an existing project. If these directories appear as children of
	// any existing projects, and are within < 2 directories removed. Then we
	// associated the var files with the project.
	for dir, files := range p.discoveredVarFiles {
		for i, project := range projects {
			if isNestedDir(project.Path, dir, 2) {
				rel, _ := filepath.Rel(project.Path, dir)

				for _, f := range files {
					projects[i].TerraformVarFiles = append(projects[i].TerraformVarFiles, filepath.Join(rel, f))
				}
			}
		}
	}

	for i := range projects {
		sort.Strings(projects[i].TerraformVarFiles)
	}

	return projects
}

// isNestedDir checks if the target path nested no more than 'levels' under the base path
func isNestedDir(basePath, targetPath string, levels int) bool {
	rel, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		return false
	}

	if strings.HasPrefix(rel, "..") || rel == "." {
		return false
	}

	// Count the separators in the relative path
	sepCount := strings.Count(rel, string(filepath.Separator))
	return sepCount <= levels
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
		p.discoveredProjects = []string{}
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

	if p.useAllPaths && len(files) > 0 {
		p.discoveredProjects = append(p.discoveredProjects, fullPath)
	} else if hasProviderBlock || hasTerraformBackendBlock {
		p.discoveredProjects = append(p.discoveredProjects, fullPath)
	}

	for _, info := range fileInfos {
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") {
				continue
			}

			p.walkPaths(filepath.Join(fullPath, info.Name()), level+1)
		}
	}

	// If it's the top level and there's Terraform files, and no other detected projects then add it as
	// a project.
	if level == 0 && len(files) > 0 && len(p.discoveredProjects) == 0 {
		p.discoveredProjects = append(p.discoveredProjects, fullPath)
	}
}
