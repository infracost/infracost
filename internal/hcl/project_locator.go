package hcl

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/sirupsen/logrus"
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
	logger         *logrus.Entry

	basePath string
}

// ProjectLocatorConfig provides configuration options on how the locator functions.
type ProjectLocatorConfig struct {
	ExcludedSubDirs []string
	ChangedObjects  []string
	UseAllPaths     bool
}

// NewProjectLocator returns safely initialized ProjectLocator.
func NewProjectLocator(logger *logrus.Entry, config *ProjectLocatorConfig) *ProjectLocator {
	if config != nil {
		return &ProjectLocator{
			modules:        make(map[string]struct{}),
			moduleCalls:    make(map[string][]string),
			excludedDirs:   config.ExcludedSubDirs,
			changedObjects: config.ChangedObjects,
			logger:         logger,
			useAllPaths:    config.UseAllPaths,
		}
	}

	return &ProjectLocator{
		modules: make(map[string]struct{}),
		logger:  logger,
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

func (p ProjectLocator) hasChanges(dir string) bool {
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
	Path string
	// HasChanges contains information about whether the project has git changes associated with it.
	// This will show as true if one or more files/directories have changed in the Path, and also if
	// and local modules that are used by this project have changes.
	HasChanges bool
}

// FindRootModules returns a list of all directories that contain a full Terraform project under the given fullPath.
// This list excludes any Terraform modules that have been found (if they have been called by a Module source).
func (p *ProjectLocator) FindRootModules(fullPath string) []RootPath {
	p.basePath, _ = filepath.Abs(fullPath)
	p.modules = make(map[string]struct{})
	p.moduleCalls = make(map[string][]string)

	isSkipped := p.buildSkippedMatcher(fullPath)
	dirs := p.walkPaths(fullPath, 0)
	p.logger.Debugf("walking directory at %s returned a list of possible Terraform projects: %+v", fullPath, dirs)

	var projects []RootPath
	for _, dir := range dirs {
		if isSkipped(dir) {
			p.logger.Debugf("skipping directory %s as it is marked as exluded by --exclude-path", dir)
			continue
		}

		if _, ok := p.modules[dir]; ok && !p.useAllPaths {
			p.logger.Debugf("skipping directory %s as it has been called as a module", dir)
			continue
		}

		projects = append(projects, RootPath{
			Path:       dir,
			HasChanges: p.hasChanges(dir),
		})
	}

	return projects
}

func (p *ProjectLocator) maxSearchDepth() int {
	if p.useAllPaths {
		return 10
	}

	return 5
}

func (p *ProjectLocator) walkPaths(fullPath string, level int) []string {
	p.logger.Debugf("walking path %s to discover terraform files", fullPath)

	if level >= p.maxSearchDepth() {
		p.logger.Debugf("exiting parsing directory %s as it is outside the maximum evaluation threshold", fullPath)
		return nil
	}

	hclParser := hclparse.NewParser()

	fileInfos, err := os.ReadDir(fullPath)
	if err != nil {
		p.logger.WithError(err).Warnf("could not get file information for path %s skipping evaluation", fullPath)
		return nil
	}

	var dirs []string
	for _, info := range fileInfos {
		if info.IsDir() {
			continue
		}

		var parseFunc func(filename string) (*hcl.File, hcl.Diagnostics)
		if strings.HasSuffix(info.Name(), ".tf") {
			parseFunc = hclParser.ParseHCLFile
		}

		if strings.HasSuffix(info.Name(), ".tf.json") {
			parseFunc = hclParser.ParseJSONFile
		}

		if parseFunc == nil {
			continue
		}

		path := filepath.Join(fullPath, info.Name())
		_, diag := parseFunc(path)
		if diag != nil && diag.HasErrors() {
			p.logger.Warnf("skipping file: %s hcl parsing err: %s", path, diag.Error())
			continue
		}
	}

	files := hclParser.Files()
	var providerBlocks bool

	for _, file := range files {
		body, content, diags := file.Body.PartialContent(justProviderBlocks)
		if diags != nil && diags.HasErrors() {
			p.logger.WithError(diags).Warnf("skipping building module information for file %s as failed to get partial body contents", file)
			continue
		}

		// only do this after looping through all files
		if len(body.Blocks) > 0 {
			providerBlocks = true
		}

		moduleBody, _, _ := content.PartialContent(justModuleBlocks)
		for _, module := range moduleBody.Blocks {
			a, _ := module.Body.JustAttributes()
			if src, ok := a["source"]; ok {
				val, _ := src.Expr.Value(nil)
				fields := logrus.Fields{
					"module": strings.Join(module.Labels, "."),
				}

				if val.Type() != cty.String {
					p.logger.WithFields(fields).Debugf("got unexpected cty value for module source string in file %s", file)
					continue
				}

				var realPath string
				err := gocty.FromCtyValue(val, &realPath)
				if err != nil {
					p.logger.WithError(err).WithFields(fields).Debug("could not read source value of module as string")
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

	// This means we're at the top level and there are Terraform files.
	// It's safe to assume that this is a Terraform project.
	if level == 0 && len(files) > 0 {
		return []string{fullPath}
	}

	if p.useAllPaths && len(files) > 0 {
		return []string{fullPath}
	}

	if providerBlocks {
		return []string{fullPath}
	}

	for _, info := range fileInfos {
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") {
				continue
			}

			childDirs := p.walkPaths(filepath.Join(fullPath, info.Name()), level+1)
			if len(childDirs) > 0 {
				dirs = append(dirs, childDirs...)
			}
		}
	}

	return dirs
}
