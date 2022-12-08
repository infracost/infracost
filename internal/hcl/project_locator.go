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
	moduleCalls  map[string]struct{}
	excludedDirs []string
	useAllPaths  bool
	logger       *logrus.Entry

	basePath string
}

// ProjectLocatorConfig provides configuration options on how the locator functions.
type ProjectLocatorConfig struct {
	ExcludedSubDirs []string
	UseAllPaths     bool
}

// NewProjectLocator returns safely initialized ProjectLocator.
func NewProjectLocator(logger *logrus.Entry, config *ProjectLocatorConfig) *ProjectLocator {
	if config != nil {
		return &ProjectLocator{
			moduleCalls:  make(map[string]struct{}),
			excludedDirs: config.ExcludedSubDirs,
			logger:       logger,
			useAllPaths:  config.UseAllPaths,
		}
	}

	return &ProjectLocator{
		moduleCalls: make(map[string]struct{}),
		logger:      logger,
	}
}

func (p *ProjectLocator) buildMatches(fullPath string) func(string) bool {
	var matches []string
	globMatches := make(map[string]struct{})

	for _, dir := range p.excludedDirs {
		var absoluteDir string
		if dir == filepath.Base(dir) {
			matches = append(matches, dir)
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
		for _, match := range matches {
			if match == base {
				return true
			}
		}

		return false
	}
}

// FindRootModules returns a list of all directories that contain a full Terraform project under the given fullPath.
// This list excludes any Terraform modules that have been found (if they have been called by a Module source).
func (p *ProjectLocator) FindRootModules(fullPath string) []string {
	p.basePath, _ = filepath.Abs(fullPath)
	p.moduleCalls = make(map[string]struct{})

	isSkipped := p.buildMatches(fullPath)
	dirs := p.walkPaths(fullPath, 0)
	p.logger.Debugf("walking directory at %s returned a list of possible Terraform projects: %+v", fullPath, dirs)

	var filtered []string
	for _, dir := range dirs {
		if isSkipped(dir) {
			p.logger.Debugf("skipping directory %s as it is marked as exluded by --exclude-path", dir)
			continue
		}

		if _, ok := p.moduleCalls[dir]; !ok {
			filtered = append(filtered, dir)
		} else {
			p.logger.Debugf("skipping directory %s as it has been called as a module", dir)
		}
	}

	return filtered
}

func (p *ProjectLocator) walkPaths(fullPath string, level int) []string {
	p.logger.Debugf("walking path %s to discover terraform files", fullPath)

	if level >= maxTfProjectSearchLevel {
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

	// if there are Terraform files at the top level then use this as the root module, no need to search for provider blocks.
	if level == 0 && len(files) > 0 {
		return []string{fullPath}
	}

	if p.useAllPaths && len(files) > 0 {
		return []string{fullPath}
	}

	for _, file := range files {
		body, content, diags := file.Body.PartialContent(justProviderBlocks)
		if diags != nil && diags.HasErrors() {
			p.logger.WithError(diags).Warnf("skipping building module information for file %s as failed to get partial body contents", file)
			continue
		}

		if len(body.Blocks) == 0 {
			continue
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
				p.moduleCalls[mp] = struct{}{}
			}
		}

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
