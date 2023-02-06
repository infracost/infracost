package hcl

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"

	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/extclient"
	"github.com/infracost/infracost/internal/hcl/modules"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/ui"
)

var (
	defaultTerraformWorkspaceName = "default"
)

type Option func(p *Parser)

// OptionWithTFVarsPaths takes a slice of paths and sets them on the parser relative
// to the Parser initialPath. Paths that don't exist will be ignored.
func OptionWithTFVarsPaths(paths []string) Option {
	return func(p *Parser) {
		var filenames []string
		for _, name := range paths {
			if path.IsAbs(name) {
				filenames = append(filenames, name)
				continue
			}

			relToProject := path.Join(p.initialPath, name)
			filenames = append(filenames, relToProject)
		}

		p.tfvarsPaths = filenames
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

// OptionWithRawCtyInput sets the input variables for the parser using a cty.Value.
// OptionWithRawCtyInput expects that this input is a ObjectValue that can be transformed to a map.
func OptionWithRawCtyInput(input cty.Value) (op Option) {
	defer func() {
		err := recover()
		if err != nil {
			logging.Logger.Debugf("recovering from panic using raw input Terraform var %s", err)
			op = func(p *Parser) {}
		}
	}()

	asMap := input.AsValueMap()

	return func(p *Parser) {
		if p.inputVars == nil {
			p.inputVars = asMap
			return
		}

		for k, v := range asMap {
			p.inputVars[k] = v
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
		p.remoteVariablesLoader = NewRemoteVariablesLoader(client, localWorkspace, p.logger, loaderOpts...)
	}
}

func OptionWithBlockBuilder(blockBuilder BlockBuilder) Option {
	return func(p *Parser) {
		p.blockBuilder = blockBuilder
	}
}

// OptionWithTerraformWorkspace informs the Parser to use the provided name as the workspace for context evaluation.
// The Parser exposes this workspace in the evaluation context under the variable named `terraform.workspace`.
// This is commonly used by users to specify different capacity/configuration in their Terraform, e.g:
//
//	terraform.workspace == "prod" ? "m5.8xlarge" : "m5.4xlarge"
func OptionWithTerraformWorkspace(name string) Option {
	name = strings.TrimSpace(name)
	return func(p *Parser) {
		if name != "" {
			if p.logger != nil {
				p.logger.Debugf("setting HCL parser to use user provided Terraform workspace: '%s'", name)
			}

			p.workspaceName = name
		}
	}
}

// OptionWithSpinner sets a SpinnerFunc onto the Parser. With this option enabled
// the Parser will send progress to the Spinner. This is disabled by default as
// we run the Parser concurrently underneath DirProvider and don't want to mess with its output.
func OptionWithSpinner(f ui.SpinnerFunc) Option {
	return func(p *Parser) {
		p.newSpinner = f

		if p.moduleLoader != nil {
			p.moduleLoader.NewSpinner = f
		}
	}
}

// Parser is a tool for parsing terraform templates at a given file system location.
type Parser struct {
	initialPath           string
	tfEnvVars             map[string]cty.Value
	defaultVarFiles       []string
	tfvarsPaths           []string
	inputVars             map[string]cty.Value
	workspaceName         string
	moduleLoader          *modules.ModuleLoader
	blockBuilder          BlockBuilder
	newSpinner            ui.SpinnerFunc
	remoteVariablesLoader *RemoteVariablesLoader
	credentialsSource     *modules.CredentialsSource
	logger                *logrus.Entry
	hasChanges            bool
}

// LoadParsers inits a list of Parser with the provided option and initialPath. LoadParsers locates Terraform files
// in the given initialPath and returns a Parser for each directory it locates a Terraform project within. If
// the initialPath contains Terraform files at the top level parsers will be len 1.
func LoadParsers(initialPath string, loader *modules.ModuleLoader, locatorConfig *ProjectLocatorConfig, logger *logrus.Entry, options ...Option) ([]*Parser, error) {
	pl := NewProjectLocator(logger, locatorConfig)
	rootPaths := pl.FindRootModules(initialPath)
	if len(rootPaths) == 0 && len(locatorConfig.ChangedObjects) > 0 {
		return nil, nil
	}

	if len(rootPaths) == 0 {
		return nil, fmt.Errorf("No valid Terraform files found at path %s, try a different directory", initialPath)
	}

	var parsers = make([]*Parser, len(rootPaths))
	for i, rootPath := range rootPaths {
		parsers[i] = newParser(rootPath, loader, logger, options...)
	}

	return parsers, nil
}

func newParser(projectRoot RootPath, moduleLoader *modules.ModuleLoader, logger *logrus.Entry, options ...Option) *Parser {
	parserLogger := logger.WithFields(logrus.Fields{
		"parser_path": projectRoot.Path,
	})

	p := &Parser{
		initialPath:   projectRoot.Path,
		hasChanges:    projectRoot.HasChanges,
		workspaceName: defaultTerraformWorkspaceName,
		blockBuilder:  BlockBuilder{SetAttributes: []SetAttributesFunc{SetUUIDAttributes}, Logger: logger},
		logger:        parserLogger,
		moduleLoader:  moduleLoader,
	}

	var defaultVarFiles []string

	defaultTfFile := path.Join(projectRoot.Path, "terraform.tfvars")
	if _, err := os.Stat(defaultTfFile); err == nil {
		parserLogger.Debugf("using terraform.tfvar file %s", defaultTfFile)
		defaultVarFiles = append(defaultVarFiles, defaultTfFile)
	}

	if _, err := os.Stat(defaultTfFile + ".json"); err == nil {
		parserLogger.Debugf("using terraform.tfvar file %s.json", defaultTfFile)
		defaultVarFiles = append(defaultVarFiles, defaultTfFile+".json")
	}

	autoVarsSuffix := ".auto.tfvars"
	infos, _ := os.ReadDir(projectRoot.Path)
	for _, info := range infos {
		name := info.Name()
		if strings.HasSuffix(name, autoVarsSuffix) || strings.HasSuffix(name, autoVarsSuffix+".json") {
			parserLogger.Debugf("using auto var file %s", name)
			defaultVarFiles = append(defaultVarFiles, path.Join(projectRoot.Path, name))
		}
	}

	p.defaultVarFiles = defaultVarFiles

	for _, option := range options {
		option(p)
	}

	return p
}

// ParseDirectory parses all the terraform files in the initalPath into Blocks and then passes them to an Evaluator
// to fill these Blocks with additional Context information. Parser does not parse any blocks outside the root Module.
// It instead leaves ModuleLoader to fetch these Modules on demand. See ModuleLoader.Load for more information.
//
// ParseDirectory returns the root Module that represents the top of the Terraform Config tree.
func (p *Parser) ParseDirectory() (m *Module, err error) {
	m = &Module{
		RootPath:   p.initialPath,
		ModulePath: p.initialPath,
	}

	defer func() {
		e := recover()
		if e != nil {
			err = clierror.NewPanicError(fmt.Errorf("%s", e), debug.Stack())
		}
	}()

	p.logger.Debugf("Beginning parse for directory '%s'...", p.initialPath)

	// load the initial root directory into a list of hcl files
	// at this point these files have no schema associated with them.
	files, err := loadDirectory(p.logger, p.initialPath, false)
	if err != nil {
		return m, err
	}

	// load the files into given hcl block types. These are then wrapped with *Block structs.
	blocks, err := p.parseDirectoryFiles(files)
	if err != nil {
		return m, err
	}

	if len(blocks) == 0 {
		return m, errors.New("No valid terraform files found given path, try a different directory")
	}

	p.logger.Debug("Loading TFVars...")
	inputVars, err := p.loadVars(blocks, p.tfvarsPaths)
	if err != nil {
		return m, err
	}

	// load the modules. This downloads any remote modules to the local file system
	modulesManifest, err := p.moduleLoader.Load(p.initialPath)
	if err != nil {
		return m, fmt.Errorf("Error loading Terraform modules: %s", err)
	}

	p.logger.Debug("Evaluating expressions...")
	workingDir, err := os.Getwd()
	if err != nil {
		return m, fmt.Errorf("Error could not evaluate current working directory %w", err)
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
		p.logger,
	)

	root, err := evaluator.Run()
	if err != nil {
		return m, err
	}

	root.HasChanges = p.hasChanges
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
			p.logger.Debugf("skipping file could not load blocks err: %s", err)
			continue
		}

		if len(fileBlocks) > 0 {
			p.logger.Debugf("Added %d blocks from %s...", len(fileBlocks), fileBlocks[0].DefRange.Filename)
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
			p.logger.Warnf("could not load vars from Terraform Cloud: %s", err)
			return combinedVars, err
		}

		for k, v := range remoteVars {
			combinedVars[k] = v
		}
	}

	for _, name := range p.defaultVarFiles {
		err := p.loadAndCombineVars(name, combinedVars)
		if err != nil {
			p.logger.WithError(err).Warnf("could not load vars from auto var file %s", name)
			continue
		}
	}

	for _, filename := range filenames {
		err := p.loadAndCombineVars(filename, combinedVars)
		if err != nil {
			return combinedVars, err
		}
	}

	for k, v := range p.inputVars {
		combinedVars[k] = v
	}

	return combinedVars, nil
}

func (p *Parser) loadAndCombineVars(filename string, combinedVars map[string]cty.Value) error {
	vars, err := p.loadVarFile(filename)
	if err != nil {
		return err
	}

	for k, v := range vars {
		combinedVars[k] = v
	}

	return nil
}

func (p *Parser) loadVarFile(filename string) (map[string]cty.Value, error) {
	inputVars := make(map[string]cty.Value)

	if filename == "" {
		return inputVars, nil
	}

	_, err := os.Stat(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("Passed var file does not exist: %s. Make sure you are passing the var file path relative to the --path flag.", filename)
		}

		return nil, fmt.Errorf("could not stat provided var file: %w", err)
	}

	var parseFunc func(filename string) (*hcl.File, hcl.Diagnostics)

	hclParser := hclparse.NewParser()

	parseFunc = hclParser.ParseHCLFile
	if strings.HasSuffix(filename, ".json") {
		parseFunc = hclParser.ParseJSONFile
	}

	variableFile, diags := parseFunc(filename)
	if diags.HasErrors() {
		p.logger.WithError(errors.New(diags.Error())).Debugf("could not parse supplied var file %s", filename)

		return inputVars, nil
	}

	attrs, _ := variableFile.Body.JustAttributes()

	fields := make(logrus.Fields, len(attrs))
	for _, attr := range attrs {
		value, diag := attr.Expr.Value(&hcl.EvalContext{})
		if diag.HasErrors() {
			p.logger.WithError(errors.New(diag.Error())).Debugf("problem evaluating input var %s", attr.Name)
		}

		inputVars[attr.Name] = value
		fields[attr.Name] = value
	}

	p.logger.WithFields(fields).Debugf("adding input vars from file %s", filename)
	return inputVars, nil
}

type file struct {
	path    string
	hclFile *hcl.File
}

func loadDirectory(logger *logrus.Entry, fullPath string, stopOnHCLError bool) ([]file, error) {
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

			logger.Debugf("skipping file: %s hcl parsing err: %s", path, diag.Error())
			continue
		}
	}

	files := make([]file, 0, len(hclParser.Files()))
	for filename, f := range hclParser.Files() {
		files = append(files, file{hclFile: f, path: filename})
	}

	return files, nil
}
