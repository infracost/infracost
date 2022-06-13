package hcl

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"

	"github.com/infracost/infracost/internal/extclient"
	"github.com/infracost/infracost/internal/hcl/modules"
	"github.com/infracost/infracost/internal/ui"
)

var (
	// This sets a global logger for this package, which is a bit of a hack. In the future we should use a context for this.
	log = logrus.StandardLogger().WithField("parser", "terraform_hcl")

	maxTfProjectSearchLevel = 5
)

type Option func(p *Parser)

// OptionWithTFVarsPaths takes a slice of paths and sets them on the parser relative
// to the Parser initialPath. Paths that don't exist will be ignored.
func OptionWithTFVarsPaths(paths []string) Option {
	return func(p *Parser) {
		var relative []string
		for _, name := range paths {
			tfvp := path.Join(p.initialPath, name)
			relative = append(relative, tfvp)
		}
		p.tfvarsPaths = relative
	}
}

func OptionStopOnHCLError() Option {
	return func(p *Parser) {
		p.stopOnHCLError = true
	}
}

// OptionWithTFEnvVars takes any TF_ENV_xxx=yyy from the environment and converts them to cty.Value
// It then sets these as the Parser starting tfEnvVars which are used at the root module evaluation.
func OptionWithTFEnvVars(projectEnv map[string]string) Option {
	return func(p *Parser) {
		ctyVars := make(map[string]cty.Value)

		// First load any TF_VARs set in the environment
		for _, v := range os.Environ() {
			if strings.HasPrefix(v, "TF_VAR_") {
				pieces := strings.Split(v[len("TF_VAR_"):], "=")
				if len(pieces) != 2 {
					continue
				}

				ctyVars[pieces[0]] = cty.StringVal(pieces[1])
			}
		}

		// Then load any TF_VARs set in the project config "env:" block
		for k, v := range projectEnv {
			if strings.HasPrefix(k, "TF_VAR_") {
				ctyVars[k[len("TF_VAR_"):]] = cty.StringVal(v)
			}
		}

		p.tfEnvVars = ctyVars
	}
}

// OptionWithPlanFlagVars takes TF var inputs specified in a command line string and converts them to cty.Value
// It sets these as the Parser starting inputVars which are used at the root module evaluation.
func OptionWithPlanFlagVars(vs []string) Option {
	return func(p *Parser) {
		if p.inputVars == nil {
			p.inputVars = make(map[string]cty.Value)
		}
		for _, v := range vs {
			pieces := strings.Split(v, "=")
			if len(pieces) != 2 {
				continue
			}

			p.inputVars[pieces[0]] = cty.StringVal(pieces[1])
		}
	}
}

// OptionWithInputVars takes cmd line var input values and converts them to cty.Value
// It sets these as the Parser starting inputVars which are used at the root module evaluation.
func OptionWithInputVars(vars map[string]string) Option {
	return func(p *Parser) {
		if p.inputVars == nil {
			p.inputVars = make(map[string]cty.Value, len(vars))
		}

		for k, v := range vars {
			p.inputVars[k] = cty.StringVal(v)
		}
	}
}

// OptionWithRemoteVarLoader accepts Terraform Cloud/Enterprise host and token
// values to load remote execution variables.
func OptionWithRemoteVarLoader(host, token, localWorkspace string) Option {
	return func(p *Parser) {
		if host == "" || token == "" {
			return
		}

		var loaderOpts []RemoteVariablesLoaderOption
		if p.newSpinner != nil {
			loaderOpts = append(loaderOpts, RemoteVariablesLoaderWithSpinner(p.newSpinner))
		}

		client := extclient.NewAuthedAPIClient(host, token)
		p.remoteVariablesLoader = NewRemoteVariablesLoader(client, localWorkspace, loaderOpts...)
	}
}

func OptionWithBlockBuilder(blockBuilder BlockBuilder) Option {
	return func(p *Parser) {
		p.blockBuilder = blockBuilder
	}
}

// OptionWithSpinner sets a SpinnerFunc onto the Parser. With this option enabled
// the Parser will send progress to the Spinner. This is disabled by default as
// we run the Parser concurrently underneath DirProvider and don't want to mess with its output.
func OptionWithSpinner(f ui.SpinnerFunc) Option {
	return func(p *Parser) {
		p.newSpinner = f
	}
}

// OptionWithWarningFunc will set the Parser writeWarning to the provided f.
// This is disabled by default as we run the Parser concurrently underneath a
// DirProvider and don't want to mess with its output.
func OptionWithWarningFunc(f ui.WriteWarningFunc) Option {
	return func(p *Parser) {
		p.writeWarning = f
	}
}

// Parser is a tool for parsing terraform templates at a given file system location.
type Parser struct {
	initialPath           string
	tfEnvVars             map[string]cty.Value
	defaultVarFiles       []string
	tfvarsPaths           []string
	inputVars             map[string]cty.Value
	stopOnHCLError        bool
	workspaceName         string
	moduleLoader          *modules.ModuleLoader
	blockBuilder          BlockBuilder
	newSpinner            ui.SpinnerFunc
	writeWarning          ui.WriteWarningFunc
	remoteVariablesLoader *RemoteVariablesLoader
}

// LoadParsers inits a list of Parser with the provided option and initialPath. LoadParsers locates Terraform files
// in the given initialPath and returns a Parser for each directory it locates a Terraform project within. If
// the initialPath contains Terraform files at the top level parsers will be len 1.
func LoadParsers(initialPath string, excludePaths []string, options ...Option) ([]*Parser, error) {
	pl := &projectLocator{moduleCalls: make(map[string]struct{}), excludedDirs: excludePaths}
	rootPaths := pl.findRootModules(initialPath)
	if len(rootPaths) == 0 {
		return nil, errors.New("No valid Terraform files found at the given path, try a different directory")
	}

	var parsers = make([]*Parser, len(rootPaths))
	for i, rootPath := range rootPaths {
		parsers[i] = newParser(rootPath, options)
	}

	return parsers, nil
}

func newParser(initialPath string, options []Option) *Parser {
	p := &Parser{
		initialPath:   initialPath,
		workspaceName: "default",
		blockBuilder:  BlockBuilder{SetAttributes: []SetAttributesFunc{SetUUIDAttributes}},
	}

	var defaultVarFiles []string

	defaultTfFile := path.Join(initialPath, "terraform.tfvars")
	if _, err := os.Stat(defaultTfFile); err == nil {
		defaultVarFiles = append(defaultVarFiles, defaultTfFile)
	}

	if _, err := os.Stat(defaultTfFile + ".json"); err == nil {
		defaultVarFiles = append(defaultVarFiles, defaultTfFile+".json")
	}

	autoVarsSuffix := ".auto.tfvars"
	infos, _ := os.ReadDir(initialPath)
	for _, info := range infos {
		name := info.Name()
		if strings.HasSuffix(name, autoVarsSuffix) || strings.HasSuffix(name, autoVarsSuffix+".json") {
			defaultVarFiles = append(defaultVarFiles, path.Join(initialPath, name))
		}
	}

	p.defaultVarFiles = defaultVarFiles

	for _, option := range options {
		option(p)
	}

	var loaderOpts []modules.LoaderOption
	if p.newSpinner != nil {
		loaderOpts = append(loaderOpts, modules.LoaderWithSpinner(p.newSpinner))
	}

	p.moduleLoader = modules.NewModuleLoader(initialPath, loaderOpts...)

	return p
}

// ParseDirectory parses all the terraform files in the initalPath into Blocks and then passes them to an Evaluator
// to fill these Blocks with additional Context information. Parser does not parse any blocks outside the root Module.
// It instead leaves ModuleLoader to fetch these Modules on demand. See ModuleLoader.Load for more information.
//
// ParseDirectory returns the root Module that represents the top of the Terraform Config tree.
func (p *Parser) ParseDirectory() (*Module, error) {
	log.Debugf("Beginning parse for directory '%s'...", p.initialPath)

	// load the initial root directory into a list of hcl files
	// at this point these files have no schema associated with them.
	files, err := loadDirectory(p.initialPath, p.stopOnHCLError)
	if err != nil {
		return nil, err
	}

	// load the files into given hcl block types. These are then wrapped with *Block structs.
	blocks, err := p.parseDirectoryFiles(files)
	if err != nil {
		return nil, err
	}

	if len(blocks) == 0 {
		return nil, errors.New("No valid terraform files found given path, try a different directory")
	}

	log.Debug("Loading TFVars...")
	inputVars, err := p.loadVars(blocks, p.tfvarsPaths)
	if err != nil {
		return nil, err
	}

	// load the modules. This downloads any remote modules to the local file system
	modulesManifest, err := p.moduleLoader.Load()
	if err != nil {
		return nil, fmt.Errorf("Error loading Terraform modules: %s", err)
	}

	log.Debug("Evaluating expressions...")
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("Error could not evaluate current working directory %w", err)
	}

	// load an Evaluator with the top level Blocks to begin Context propagation.
	evaluator := NewEvaluator(
		Module{
			Name:       "",
			Source:     "",
			Blocks:     blocks,
			RawBlocks:  blocks,
			RootPath:   p.initialPath,
			ModulePath: p.initialPath,
		},
		workingDir,
		inputVars,
		modulesManifest,
		nil,
		p.workspaceName,
		p.blockBuilder,
		p.newSpinner,
	)

	if v := evaluator.MissingVars(); len(v) > 0 {
		if p.writeWarning != nil {
			p.writeWarning(
				fmt.Sprintf(
					"Input values were not provided for following Terraform variables: %s. %s",
					strings.TrimRight(strings.Join(v, ", "), ", "),
					"Use --terraform-var-file or --terraform-var to specify them.",
				),
			)
		}
	}

	root, err := evaluator.Run()
	if err != nil {
		return nil, err
	}

	return root, nil
}

// Path returns the full path that the parser runs within.
func (p *Parser) Path() string {
	return p.initialPath
}

func (p *Parser) parseDirectoryFiles(files []file) (Blocks, error) {
	var blocks Blocks

	for _, file := range files {
		fileBlocks, err := loadBlocksFromFile(file, nil)
		if err != nil {
			if p.stopOnHCLError {
				return nil, err
			}

			log.Warnf("skipping file could not load blocks err: %s", err)
			continue
		}

		if len(fileBlocks) > 0 {
			log.Debugf("Added %d blocks from %s...", len(fileBlocks), fileBlocks[0].DefRange.Filename)
		}

		for _, fileBlock := range fileBlocks {
			blocks = append(
				blocks,
				p.blockBuilder.NewBlock(file.path, fileBlock, nil, nil),
			)
		}
	}

	return blocks, nil
}

func (p *Parser) loadVars(blocks Blocks, filenames []string) (map[string]cty.Value, error) {
	combinedVars := p.tfEnvVars
	if combinedVars == nil {
		combinedVars = make(map[string]cty.Value)
	}

	if p.remoteVariablesLoader != nil {
		remoteVars, err := p.remoteVariablesLoader.Load(blocks)

		if err != nil {
			log.Warnf("could not load vars from Terraform Cloud: %s", err)
			return combinedVars, err
		}

		for k, v := range remoteVars {
			combinedVars[k] = v
		}
	}

	for _, name := range p.defaultVarFiles {
		err := loadAndCombineVars(name, combinedVars)
		if err != nil {
			log.Warnf("could not load vars from auto var file %s err: %s", name, err)
			continue
		}
	}

	for _, filename := range filenames {
		err := loadAndCombineVars(filename, combinedVars)
		if err != nil {
			return combinedVars, err
		}
	}

	for k, v := range p.inputVars {
		combinedVars[k] = v
	}

	return combinedVars, nil
}

func loadAndCombineVars(filename string, combinedVars map[string]cty.Value) error {
	vars, err := loadVarFile(filename)
	if err != nil {
		return err
	}

	for k, v := range vars {
		combinedVars[k] = v
	}

	return nil
}

func loadVarFile(filename string) (map[string]cty.Value, error) {
	inputVars := make(map[string]cty.Value)

	if filename == "" {
		return inputVars, nil
	}

	_, err := os.Stat(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("Passed var file does not exist: %s", filename)
		}
		return nil, fmt.Errorf("Could not stat var file: %s", err)
	}

	log.Debugf("loading tfvars-file [%s]", filename)
	src, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Could not read var file file %s %w", filename, err)
	}

	variableFile, _ := hclsyntax.ParseConfig(src, filename, hcl.Pos{Line: 1, Column: 1})
	attrs, _ := variableFile.Body.JustAttributes()

	for _, attr := range attrs {
		log.Debugf("Setting '%s' from tfvars file at %s", attr.Name, filename)
		inputVars[attr.Name], _ = attr.Expr.Value(&hcl.EvalContext{})
	}

	return inputVars, nil
}

type file struct {
	path    string
	hclFile *hcl.File
}

func loadDirectory(fullPath string, stopOnHCLError bool) ([]file, error) {
	hclParser := hclparse.NewParser()

	fileInfos, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

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

		// this is not a file we can parse:
		if parseFunc == nil {
			continue
		}

		path := filepath.Join(fullPath, info.Name())
		_, diag := parseFunc(path)
		if diag != nil && diag.HasErrors() {
			if stopOnHCLError {
				return nil, diag
			}

			log.Warnf("skipping file: %s hcl parsing err: %s", path, diag.Error())
			continue
		}
	}

	files := make([]file, 0, len(hclParser.Files()))
	for filename, f := range hclParser.Files() {
		files = append(files, file{hclFile: f, path: filename})
	}

	return files, nil
}

type projectLocator struct {
	moduleCalls  map[string]struct{}
	excludedDirs []string
}

func (p *projectLocator) buildMatches(fullPath string) func(string) bool {
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

func (p *projectLocator) findRootModules(fullPath string) []string {
	isSkipped := p.buildMatches(fullPath)
	dirs := p.walkPaths(fullPath, 0)

	var filtered []string
	for _, dir := range dirs {
		if isSkipped(dir) {
			log.Debugf("skipping directory %s as it is marked as exluded by --exclude-path", dir)
			continue
		}

		if _, ok := p.moduleCalls[dir]; !ok {
			filtered = append(filtered, dir)
		}
	}

	return filtered
}

func (p *projectLocator) walkPaths(fullPath string, level int) []string {
	if level >= maxTfProjectSearchLevel {
		return nil
	}

	hclParser := hclparse.NewParser()

	fileInfos, err := os.ReadDir(fullPath)
	if err != nil {
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
			log.Warnf("skipping file: %s hcl parsing err: %s", path, diag.Error())
			continue
		}
	}

	files := hclParser.Files()

	// if there are Terraform files at the top level then use this as the root module, no need to search for provider blocks.
	if level == 0 && len(files) > 0 {
		return []string{fullPath}
	}

	for _, file := range files {
		body, content, diags := file.Body.PartialContent(justProviderBlocks)
		if diags != nil && diags.HasErrors() {
			continue
		}

		if len(body.Blocks) > 0 {
			moduleBody, _, _ := content.PartialContent(justModuleBlocks)
			for _, module := range moduleBody.Blocks {
				a, _ := module.Body.JustAttributes()
				if src, ok := a["source"]; ok {
					val, _ := src.Expr.Value(nil)
					if val.Type() == cty.String {
						realPath := filepath.Join(fullPath, val.AsString())
						p.moduleCalls[realPath] = struct{}{}
					}
				}
			}

			return []string{fullPath}
		}
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
