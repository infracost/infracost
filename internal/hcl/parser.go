package hcl

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/extclient"
	"github.com/infracost/infracost/internal/hcl/modules"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/ui"
)

var (
	defaultTerraformWorkspaceName = "default"
	// globalTerraformVarNames is a list of var file naming convention that suggests they are applied
	// to every project, despite changes in environment.
	globalTerraformVarNames = []string{
		"default",
		"defaults",
		"global",
		"globals",
		"shared",
	}

	envPrefixRegxp = regexp.MustCompile(`^\w+-`)
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

// OptionWithModuleSuffix sets an optional module suffix which will be added to the Module after it has finished parsing
// this can be used to augment auto-detected project path names and metadata.
func OptionWithModuleSuffix(suffix string) Option {
	return func(p *Parser) {
		p.moduleSuffix = suffix
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
			logging.Logger.Debug().Msgf("recovering from panic using raw input Terraform var %s", err)
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
			p.logger.Debug().Msgf("setting HCL parser to use user provided Terraform workspace: '%s'", name)

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
	repoPath              string
	initialPath           string
	tfEnvVars             map[string]cty.Value
	defaultVarFiles       []string
	tfvarsPaths           []string
	inputVars             map[string]cty.Value
	workspaceName         string
	moduleLoader          *modules.ModuleLoader
	hclParser             *modules.SharedHCLParser
	blockBuilder          BlockBuilder
	newSpinner            ui.SpinnerFunc
	remoteVariablesLoader *RemoteVariablesLoader
	logger                zerolog.Logger
	hasChanges            bool
	moduleSuffix          string
}

// LoadParsers inits a list of Parser with the provided option and initialPath. LoadParsers locates Terraform files
// in the given initialPath and returns a Parser for each directory it locates a Terraform project within. If
// the initialPath contains Terraform files at the top level parsers will be len 1.
func LoadParsers(ctx *config.ProjectContext, initialPath string, loader *modules.ModuleLoader, locatorConfig *ProjectLocatorConfig, logger zerolog.Logger, options ...Option) ([]*Parser, error) {
	var rootPaths []RootPath
	if locatorConfig != nil && locatorConfig.SkipAutoDetection {
		rootPaths = []RootPath{
			{
				Path: initialPath,
			},
		}
	} else {
		pl := NewProjectLocator(logger, locatorConfig)
		rootPaths = pl.FindRootModules(initialPath)
	}
	if len(rootPaths) == 0 && len(locatorConfig.ChangedObjects) > 0 {
		return nil, nil
	}

	if len(rootPaths) == 0 {
		return nil, fmt.Errorf("No valid Terraform files found at path %s, try a different directory", initialPath)
	}

	if ctx.RunContext.Config.ConfigFilePath == "" && len(ctx.ProjectConfig.TerraformVarFiles) == 0 {
		var parsers []*Parser
		for _, rootPath := range rootPaths {

			var varFiles []string
			var autoVarFiles []string

			// first remove all the "auto" tfvar files from the list of discovered var files.
			// These files should not constitute a new "project" as they won't define an
			// environment but defaults that should be applied across all environments.
			for _, varFile := range rootPath.TerraformVarFiles {
				withoutJSONSuffix := strings.TrimSuffix(varFile, ".json")
				if strings.HasSuffix(withoutJSONSuffix, ".auto.tfvars") || withoutJSONSuffix == "terraform.tfvars" {
					autoVarFiles = append(autoVarFiles, varFile)
					continue
				}

				global := false

				for _, name := range globalTerraformVarNames {
					// check if the var file is a "global" one, this is only applicable if we match a
					// globalTerraformVarName and these don't have an environment prefix e.g.
					// defaults.tfvars, global.tfvars are applicable, prod-default.tfvars,
					// stag-globals are not.
					if strings.HasSuffix(withoutJSONSuffix, name+".tfvars") && !envPrefixRegxp.MatchString(withoutJSONSuffix) {
						autoVarFiles = append(autoVarFiles, varFile)
						global = true
						continue
					}
				}

				if global {
					continue
				}

				varFiles = append(varFiles, varFile)
			}

			// if we have more than 1 var file we should split the projects by var file because there is a high
			// likelihood that these var files indicate different environments/configuration.
			if len(varFiles) > 1 {
				sort.Strings(rootPath.TerraformVarFiles)

				for _, varFile := range varFiles {
					parsers = append(parsers, newParser(rootPath, loader, logger, append(
						options,
						OptionWithTFVarsPaths(append(autoVarFiles, varFile)),
						OptionWithModuleSuffix(strings.TrimSuffix(strings.TrimSuffix(varFile, ".json"), ".tfvars")),
					)...))
				}

				continue
			}

			parserOpts := options
			if len(varFiles) == 1 || len(autoVarFiles) > 0 {
				parserOpts = append(parserOpts, OptionWithTFVarsPaths(append(varFiles, autoVarFiles...)))
			}

			parsers = append(parsers, newParser(rootPath, loader, logger, parserOpts...))
		}

		return parsers, nil
	}

	var parsers = make([]*Parser, len(rootPaths))
	for i, rootPath := range rootPaths {
		parsers[i] = newParser(rootPath, loader, logger, options...)
	}

	return parsers, nil
}

func newParser(projectRoot RootPath, moduleLoader *modules.ModuleLoader, logger zerolog.Logger, options ...Option) *Parser {
	parserLogger := logger.With().Str(
		"parser_path", projectRoot.Path,
	).Logger()

	hclParser := modules.NewSharedHCLParser()

	p := &Parser{
		repoPath:      projectRoot.RepoPath,
		initialPath:   projectRoot.Path,
		hasChanges:    projectRoot.HasChanges,
		workspaceName: defaultTerraformWorkspaceName,
		hclParser:     hclParser,
		blockBuilder:  BlockBuilder{SetAttributes: []SetAttributesFunc{SetUUIDAttributes}, Logger: logger, HCLParser: hclParser},
		logger:        parserLogger,
		moduleLoader:  moduleLoader,
	}

	var defaultVarFiles []string

	defaultTfFile := path.Join(projectRoot.Path, "terraform.tfvars")
	if _, err := os.Stat(defaultTfFile); err == nil {
		parserLogger.Debug().Msgf("using terraform.tfvar file %s", defaultTfFile)
		defaultVarFiles = append(defaultVarFiles, defaultTfFile)
	}

	if _, err := os.Stat(defaultTfFile + ".json"); err == nil {
		parserLogger.Debug().Msgf("using terraform.tfvar file %s.json", defaultTfFile)
		defaultVarFiles = append(defaultVarFiles, defaultTfFile+".json")
	}

	autoVarsSuffix := ".auto.tfvars"
	infos, _ := os.ReadDir(projectRoot.Path)
	for _, info := range infos {
		name := info.Name()
		if strings.HasSuffix(name, autoVarsSuffix) || strings.HasSuffix(name, autoVarsSuffix+".json") {
			parserLogger.Debug().Msgf("using auto var file %s", name)
			defaultVarFiles = append(defaultVarFiles, path.Join(projectRoot.Path, name))
		}
	}

	p.defaultVarFiles = defaultVarFiles

	for _, option := range options {
		option(p)
	}

	return p
}

// YAML returns a yaml representation of Parser, that can be used to "explain" the auto-detection functionality.
func (p *Parser) YAML() string {
	str := strings.Builder{}

	str.WriteString(fmt.Sprintf("  - path: %s\n    name: %s\n", p.RelativePath(), p.ProjectName()))
	if len(p.TerraformVarFiles()) > 0 {
		str.WriteString("    terraform_var_files:\n")

		for _, varFile := range p.TerraformVarFiles() {
			str.WriteString(fmt.Sprintf("      - %s\n", varFile))
		}
	}

	str.WriteString(fmt.Sprintf("    terraform_workspace: %s\n", p.workspaceName))

	if len(p.tfEnvVars) > 0 {
		str.WriteString("  terraform_vars:\n")

		keys := make([]string, len(p.tfEnvVars), 0)
		for key := range p.tfEnvVars {
			keys = append(keys, key)
		}

		sort.Strings(keys)

		for _, key := range keys {
			var value string
			err := gocty.FromCtyValue(p.tfEnvVars[key], &value)
			if err != nil {
				continue
			}

			str.WriteString(fmt.Sprintf("      %s:%s\n", key, value))
		}
	}

	return str.String()
}

// ParseDirectory parses all the terraform files in the initialPath into Blocks and then passes them to an Evaluator
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

	p.logger.Debug().Msgf("Beginning parse for directory '%s'...", p.initialPath)

	// load the initial root directory into a list of hcl files
	// at this point these files have no schema associated with them.
	files, err := loadDirectory(p.hclParser, p.logger, p.initialPath, false)
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

	p.logger.Debug().Msg("Loading TFVars...")
	inputVars, err := p.loadVars(blocks, p.tfvarsPaths)
	if err != nil {
		return m, err
	}

	// load the modules. This downloads any remote modules to the local file system
	modulesManifest, err := p.moduleLoader.Load(p.initialPath)
	if err != nil {
		return m, fmt.Errorf("Error loading Terraform modules: %w", err)
	}

	p.logger.Debug().Msg("Evaluating expressions...")
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
		nil,
	)

	var root *Module

	// Graph evaluation
	if evaluator.isGraphEvaluator() {
		// we use the base zerolog log here so that it's consistent with the spinner logs
		log.Info().Msgf("Building project with experimental graph runner")

		g, err := NewGraphWithRoot(p.logger, nil)
		if err != nil {
			return m, err
		}

		root, err = g.Run(evaluator)
		if err != nil {
			return m, err
		}
	} else {
		// Existing evaluation
		root, err = evaluator.Run()
		if err != nil {
			return m, err
		}
	}

	root.HasChanges = p.hasChanges
	root.TerraformVarsPaths = p.tfvarsPaths
	root.ModuleSuffix = p.moduleSuffix
	return root, nil
}

// Path returns the full path that the parser runs within.
func (p *Parser) Path() string {
	return p.initialPath
}

// RelativePath returns the path of the parser relative to the repo path
func (p *Parser) RelativePath() string {
	r, _ := filepath.Rel(p.repoPath, p.initialPath)
	return r
}

// ProjectName generates a name for the project that can be used
// in the Infracost config file.
func (p *Parser) ProjectName() string {
	r := p.RelativePath()
	name := strings.TrimSuffix(r, "/")
	name = strings.ReplaceAll(name, "/", "-")
	if p.moduleSuffix != "" {
		name = fmt.Sprintf("%s-%s", name, p.moduleSuffix)
	}

	return name
}

// TerraformVarFiles returns the list of terraform var files that the parser
// will use to load variables from.
func (p *Parser) TerraformVarFiles() []string {
	varFilesMap := make(map[string]struct{}, len(p.defaultVarFiles)+len(p.tfvarsPaths))
	varFiles := make([]string, 0, len(p.defaultVarFiles)+len(p.tfvarsPaths))

	for _, varFile := range p.defaultVarFiles {
		p, err := filepath.Rel(p.initialPath, varFile)
		if err != nil {
			continue
		}

		if _, ok := varFilesMap[p]; !ok {
			varFilesMap[p] = struct{}{}
			varFiles = append(varFiles, p)
		}
	}

	for _, varFile := range p.tfvarsPaths {
		p, err := filepath.Rel(p.initialPath, varFile)
		if err != nil {
			continue
		}

		if _, ok := varFilesMap[p]; !ok {
			varFilesMap[p] = struct{}{}
			varFiles = append(varFiles, p)
		}
	}

	return varFiles
}

func (p *Parser) parseDirectoryFiles(files []file) (Blocks, error) {
	var blocks Blocks

	for _, file := range files {
		fileBlocks, err := loadBlocksFromFile(file, nil)
		if err != nil {
			p.logger.Debug().Msgf("skipping file could not load blocks err: %s", err)
			continue
		}

		if len(fileBlocks) > 0 {
			p.logger.Debug().Msgf("Added %d blocks from %s...", len(fileBlocks), fileBlocks[0].DefRange.Filename)
		}

		for _, fileBlock := range fileBlocks {
			blocks = append(
				blocks,
				p.blockBuilder.NewBlock(file.path, p.initialPath, fileBlock, nil, nil, nil),
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
			p.logger.Warn().Msgf("could not load vars from Terraform Cloud: %s", err)
			return combinedVars, err
		}

		for k, v := range remoteVars {
			combinedVars[k] = v
		}
	}

	for _, name := range p.defaultVarFiles {
		err := p.loadAndCombineVars(name, combinedVars)
		if err != nil {
			p.logger.Warn().Err(err).Msgf("could not load vars from auto var file %s", name)
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

	parseFunc = p.hclParser.ParseHCLFile
	if strings.HasSuffix(filename, ".json") {
		parseFunc = p.hclParser.ParseJSONFile
	}

	variableFile, diags := parseFunc(filename)
	if diags.HasErrors() {
		// Check the diagnostics aren't all just Attribute Errors.
		// We can safely ignore these Attribute errors and continue to get the raw attributes.
		// The first value for the variable will be used.
		if areDiagnosticsAttributeErrors(diags) {
			p.logger.Debug().Err(errors.New(diags.Error())).Msgf("duplicate variables detected parsing file %s, using the first values defined", filename)
		} else {
			p.logger.Debug().Err(errors.New(diags.Error())).Msgf("could not parse supplied var file %s", filename)

			return inputVars, nil
		}
	}

	attrs, _ := variableFile.Body.JustAttributes()

	for _, attr := range attrs {
		value, diag := attr.Expr.Value(&hcl.EvalContext{})
		if diag.HasErrors() {
			p.logger.Debug().Err(errors.New(diag.Error())).Msgf("problem evaluating input var %s", attr.Name)
		}

		inputVars[attr.Name] = value
	}

	return inputVars, nil
}

func areDiagnosticsAttributeErrors(diags hcl.Diagnostics) bool {
	for _, diag := range diags {
		if diag.Summary != "Attribute redefined" {
			return false
		}
	}

	return true
}

type file struct {
	path    string
	hclFile *hcl.File
}

func loadDirectory(hclParser *modules.SharedHCLParser, logger zerolog.Logger, fullPath string, stopOnHCLError bool) ([]file, error) {
	fileInfos, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	files := make([]file, 0)

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
		f, diag := parseFunc(path)
		if diag != nil && diag.HasErrors() {
			if stopOnHCLError {
				return nil, diag
			}

			logger.Debug().Msgf("skipping file: %s hcl parsing err: %s", path, diag.Error())
			continue
		}

		files = append(files, file{hclFile: f, path: path})
	}

	// sort files by path to ensure consistent ordering
	sort.Slice(files, func(i, j int) bool {
		return files[i].path < files[j].path
	})

	return files, nil
}
