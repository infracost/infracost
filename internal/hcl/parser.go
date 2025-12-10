package hcl

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/rs/zerolog"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"golang.org/x/exp/maps"

	"github.com/infracost/infracost/internal/metrics"

	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/extclient"
	"github.com/infracost/infracost/internal/hcl/modules"
	"github.com/infracost/infracost/internal/logging"
)

var defaultTerraformWorkspaceName = "default"

type Option func(p *Parser)

// OptionWithTFVarsPaths takes a slice of paths adds them to the parser tfvar
// files relative to the Parser detectedProjectPath. It sorts tfvar paths for precedence
// before adding them to the parser. Paths that don't exist will be ignored.
func OptionWithTFVarsPaths(paths []string, autoDetected bool) Option {
	return func(p *Parser) {
		filenames := makePathsRelativeToInitial(paths, p.detectedProjectPath)
		tfVarsPaths := p.tfvarsPaths
		tfVarsPaths = append(tfVarsPaths, filenames...)
		p.sortVarFilesByPrecedence(tfVarsPaths, autoDetected)

		p.tfvarsPaths = tfVarsPaths
	}
}

// sortVarFilesByPrecedence sorts the given Terraform var files by the order of precedence
// as defined by Terraform. See https://www.terraform.io/docs/language/values/variables.html#variable-definition-precedence
// for more information.
//
// 1. terraform.tfvars
// 2. terraform.tfvars.json
// 3. *.auto.tfvars or *.auto.tfvars.json files, processed in lexical order of their filenames.
// 4. all other terraform var files
func (p *Parser) sortVarFilesByPrecedence(paths []string, autoDetected bool) {
	sort.Slice(paths, func(i, j int) bool {
		countI := strings.Count(paths[i], string(filepath.Separator))
		countJ := strings.Count(paths[j], string(filepath.Separator))

		getPrecedence := func(path string) int {
			switch {
			case strings.HasSuffix(path, "terraform.tfvars"):
				return 1
			case strings.HasSuffix(path, "terraform.tfvars.json"):
				return 2
			case strings.HasSuffix(path, ".auto.tfvars"), strings.HasSuffix(path, ".auto.tfvars.json"):
				return 3
			case autoDetected:
				if p.envMatcher.IsGlobalVarFile(path) {
					return 4
				}

				return 5
			default:
				return 4
			}
		}

		if autoDetected && countI != countJ {
			return countI < countJ
		}

		// Compare the precedence of two paths
		precedenceI := getPrecedence(paths[i])
		precedenceJ := getPrecedence(paths[j])

		// If they have the same precedence, sort alphabetically
		if precedenceI == precedenceJ {
			return paths[i] > paths[j]
		}

		// Otherwise, sort by precedence
		return precedenceI < precedenceJ
	})
}

func makePathsRelativeToInitial(paths []string, initialPath string) []string {

	filenames := make([]string, 0, len(paths))
	for _, name := range paths {
		if path.IsAbs(name) {
			filenames = append(filenames, name)
			continue
		}

		relToProject := path.Join(initialPath, name)
		filenames = append(filenames, relToProject)
	}

	return filenames
}

// OptionWithModuleSuffix sets an optional module suffix which will be added to the Module after it has finished parsing
// this can be used to augment auto-detected project path names and metadata. If the suffix is already part of the project name - ignore it.
func OptionWithModuleSuffix(rootPath, suffix string) Option {
	return func(p *Parser) {
		pathEnv := ""
		if p.envMatcher != nil {
			pathEnv = p.envMatcher.PathEnv(rootPath)
		}

		if pathEnv == "" || (suffix != "" && pathEnv != suffix) {
			p.moduleSuffix = suffix
		}
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
func OptionWithInputVars(vars map[string]interface{}) Option {
	return func(p *Parser) {
		if p.inputVars == nil {
			p.inputVars = make(map[string]cty.Value, len(vars))
		}

		for k, val := range vars {
			v, err := ParseVariable(val)
			if err != nil {
				p.logger.Debug().Err(err).Msgf("could not parse input var %s", k)
				continue
			}

			p.inputVars[k] = v
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

// OptionWithTFCRemoteVarLoader accepts Terraform Cloud/Enterprise host and token
// values to load remote execution variables.
func OptionWithTFCRemoteVarLoader(host, token, localWorkspace string, loaderOpts ...TFCRemoteVariablesLoaderOption) Option {
	return func(p *Parser) {
		if host == "" || token == "" {
			return
		}

		client := extclient.NewAuthedAPIClient(host, token)
		p.remoteVariableLoaders = append(p.remoteVariableLoaders, NewTFCRemoteVariablesLoader(client, localWorkspace, p.logger, loaderOpts...))
	}
}

// OptionWithSpaceliftRemoteVarLoader attempts to build a SpaceLift remote
// variable loader and set this on the parser. If the required environment
// variables are not set, the loader is skipped.
func OptionWithSpaceliftRemoteVarLoader(ctx *config.ProjectContext) Option {
	return func(p *Parser) {
		if ctx.ProjectConfig.SpaceliftAPIKeyEndpoint == "" || ctx.ProjectConfig.SpaceliftAPIKeyID == "" || ctx.ProjectConfig.SpaceliftAPIKeySecret == "" {
			p.logger.Trace().Msg("Required Spacelift API key environment variables not set, skipping Spacelift remote variable loader.")
			return
		}

		loader, err := NewSpaceliftRemoteVariableLoader(
			ctx.RunContext.VCSMetadata,
			ctx.ProjectConfig.SpaceliftAPIKeyEndpoint,
			ctx.ProjectConfig.SpaceliftAPIKeyID,
			ctx.ProjectConfig.SpaceliftAPIKeySecret,
		)
		if err != nil {
			p.logger.Debug().Err(err).Msg("could not create Spacelift remote variable loader")
			return
		}

		p.remoteVariableLoaders = append(p.remoteVariableLoaders, loader)
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

// OptionGraphEvaluator sets the Parser to use the experimental graph evaluator.
func OptionGraphEvaluator() Option {
	return func(p *Parser) {
		p.isGraph = true
		p.blockBuilder.isGraph = true
	}
}

// OptionWithProjectName sets the project name for the parser.
// This is used if the project name has been explicitly set by the user or the autodetection
func OptionWithProjectName(name string) Option {
	return func(p *Parser) {
		p.projectName = name
	}
}

type DetectedProject interface {
	ProjectName() string
	EnvName() string
	RelativePath() string
	VarFiles() []string
	YAML() string
	DependencyPaths() []string
}

// Parser is a tool for parsing terraform templates at a given file system location.
type Parser struct {
	startingPath          string
	detectedProjectPath   string
	projectName           string
	tfEnvVars             map[string]cty.Value
	tfvarsPaths           []string
	inputVars             map[string]cty.Value
	workspaceName         string
	moduleLoader          *modules.ModuleLoader
	hclParser             *modules.SharedHCLParser
	blockBuilder          BlockBuilder
	remoteVariableLoaders []RemoteVariableLoader
	logger                zerolog.Logger
	isGraph               bool
	hasChanges            bool
	moduleSuffix          string
	envMatcher            *EnvFileMatcher
	moduleCalls           []string
}

// NewParser creates a new parser for the given RootPath.
func NewParser(projectRoot RootPath, envMatcher *EnvFileMatcher, moduleLoader *modules.ModuleLoader, logger zerolog.Logger, options ...Option) *Parser {
	parserLogger := logger.With().Str(
		"parser_path", projectRoot.DetectedPath,
	).Logger()

	hclParser := modules.NewSharedHCLParser()

	p := &Parser{
		startingPath:        projectRoot.StartingPath,
		detectedProjectPath: projectRoot.DetectedPath,
		hasChanges:          projectRoot.HasChanges,
		moduleCalls:         projectRoot.ModuleCalls,
		hclParser:           hclParser,
		blockBuilder:        BlockBuilder{SetAttributes: []SetAttributesFunc{SetUUIDAttributes}, Logger: logger, HCLParser: hclParser},
		logger:              parserLogger,
		moduleLoader:        moduleLoader,
		envMatcher:          envMatcher,
	}

	for _, option := range options {
		option(p)
	}

	// if the project name is not set by the user, we will use the detected env name
	// as the project name.
	if p.workspaceName == "" && p.ProjectName() != p.EnvName() {
		p.workspaceName = p.EnvName()
	}

	return p
}

// YAML returns a yaml representation of Parser, that can be used to "explain" the auto-detection functionality.
func (p *Parser) YAML() string {
	str := strings.Builder{}

	str.WriteString(fmt.Sprintf("  - path: %s\n    name: %s\n", p.RelativePath(), p.ProjectName()))
	if len(p.VarFiles()) > 0 {
		str.WriteString("    terraform_var_files:\n")

		for _, varFile := range p.VarFiles() {
			str.WriteString(fmt.Sprintf("      - %s\n", varFile))
		}
	}

	str.WriteString("    skip_autodetect: true\n")

	if env := p.EnvName(); env != "" && env != p.ProjectName() {
		str.WriteString(fmt.Sprintf("    terraform_workspace: %s\n", env))
	}

	if len(p.tfEnvVars) > 0 {
		str.WriteString("    terraform_vars:\n")

		keys := make([]string, 0, len(p.tfEnvVars))
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

	depPaths := p.DependencyPaths()
	if len(depPaths) > 0 {
		str.WriteString("    dependency_paths:\n")

		for _, depFile := range depPaths {
			str.WriteString(fmt.Sprintf("      - %s\n", depFile))
		}
	}

	return str.String()
}

// DependencyPaths returns the list of module calls that the project depends on.
// These are usually local module calls that have been detected in the project.
func (p *Parser) DependencyPaths() []string {
	if len(p.moduleCalls) == 0 && len(p.tfvarsPaths) == 0 {
		return nil
	}

	sortedCalls := make([]string, 0, len(p.moduleCalls)+len(p.tfvarsPaths)+1)
	for _, call := range p.moduleCalls {
		relCall := call
		dep, err := filepath.Rel(p.startingPath, call)
		if err == nil {
			relCall = dep
		}

		sortedCalls = append(sortedCalls, relCall+"/**")
	}

	for _, varFile := range p.tfvarsPaths {
		relVarFile, err := filepath.Rel(p.startingPath, varFile)
		if err == nil {
			sortedCalls = append(sortedCalls, relVarFile)
		}
	}

	if p.RelativePath() == "." {
		sortedCalls = append(sortedCalls, `"**"`)
	} else {
		sortedCalls = append(sortedCalls, p.RelativePath()+"/**")
	}
	return p.uniqueDependencyPaths(sortedCalls)
}

func (p *Parser) uniqueDependencyPaths(paths []string) []string {
	seen := make(map[string]struct{})

	for _, path := range paths {
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
	}
	deps := maps.Keys(seen)
	sort.Strings(deps)
	return deps
}

// ParseDirectory parses all the terraform files in the detectedProjectPath into Blocks and then passes them to an Evaluator
// to fill these Blocks with additional Context information. Parser does not parse any blocks outside the root Module.
// It instead leaves ModuleLoader to fetch these Modules on demand. See ModuleLoader.Load for more information.
//
// ParseDirectory returns the root Module that represents the top of the Terraform Config tree.
func (p *Parser) ParseDirectory() (m *Module, err error) {
	m = &Module{
		RootPath:     p.detectedProjectPath,
		ModulePath:   p.detectedProjectPath,
		ModuleSuffix: p.moduleSuffix,
	}

	defer func() {
		e := recover()
		if e != nil {
			err = clierror.NewPanicError(fmt.Errorf("%s", e), debug.Stack())
		}
	}()

	p.logger.Debug().Msgf("Beginning parse for directory '%s'...", p.detectedProjectPath)

	metrics.GetCounter("hcl_file.count", true).Inc()

	// load the initial root directory into a list of hcl files
	// at this point these files have no schema associated with them.
	hclTimer := metrics.GetTimer("hcl_file.load.duration", false, p.detectedProjectPath).Start()
	files, err := loadDirectory(p.hclParser, p.logger, p.detectedProjectPath, false)
	hclTimer.Stop()
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

	metrics.GetCounter("block.count", true).Add(len(blocks))

	p.logger.Debug().Msg("Loading TFVars...")
	varLoadTimer := metrics.GetTimer("var.load.duration", false, p.detectedProjectPath).Start()
	inputVars, err := p.loadVars(blocks, p.tfvarsPaths)
	varLoadTimer.Stop()
	if err != nil {
		return m, err
	}

	// load the modules. This downloads any remote modules to the local file system
	modulesManifest, err := p.moduleLoader.Load(p.detectedProjectPath)
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
			RootPath:   p.detectedProjectPath,
			ModulePath: p.detectedProjectPath,
		},
		workingDir,
		inputVars,
		modulesManifest,
		nil,
		p.workspaceName,
		p.blockBuilder,
		p.logger,
		p.isGraph,
	)

	var root *Module

	evaluationTimer := metrics.GetTimer("eval.duration", false, p.detectedProjectPath).Start()
	defer evaluationTimer.Stop()

	// Graph evaluation
	if evaluator.isGraph {
		logging.Logger.Debug().Msg("Building project with experimental graph runner")

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

	// Set the root module's provider constraints we only look to do this at the root
	// module level for now but we might look to change this in the future to parse
	// on a per-module basis, with the same level of inheritance as the provider
	// blocks. However, for our current use case this is not really necessary, and
	// would just add complexity.
	root.ProviderConstraints = NewProviderConstraints(blocks)

	root.HasChanges = p.hasChanges
	root.TerraformVarsPaths = p.tfvarsPaths
	root.ModuleSuffix = p.moduleSuffix
	return root, nil
}

// Path returns the full path that the parser runs within.
func (p *Parser) Path() string {
	return p.detectedProjectPath
}

// RelativePath returns the path of the parser relative to the repo path
func (p *Parser) RelativePath() string {
	r, _ := filepath.Rel(p.startingPath, p.detectedProjectPath)
	return r
}

// ProjectName generates a name for the project that can be used
// in the Infracost config file.
func (p *Parser) ProjectName() string {
	if p.projectName != "" {
		return p.projectName
	}

	name := config.CleanProjectName(p.RelativePath())

	if p.moduleSuffix != "" {
		name = fmt.Sprintf("%s-%s", name, p.moduleSuffix)
	} else if p.workspaceName != "" && p.workspaceName != defaultTerraformWorkspaceName && !strings.HasSuffix(name, p.workspaceName) {
		name = fmt.Sprintf("%s-%s", name, p.workspaceName)
	}

	return name
}

// EnvName returns the module suffix of the parser (normally the environment name).
func (p *Parser) EnvName() string {
	if p.moduleSuffix != "" {
		return p.moduleSuffix
	}

	return p.envMatcher.EnvName(p.ProjectName())
}

// TerraformVarFiles returns the list of terraform var files that the parser
// will use to load variables from.
func (p *Parser) VarFiles() []string {
	varFilesMap := make(map[string]struct{}, len(p.tfvarsPaths))
	varFiles := make([]string, 0, len(p.tfvarsPaths))

	for _, varFile := range p.tfvarsPaths {
		p, err := filepath.Rel(p.detectedProjectPath, varFile)
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
				p.blockBuilder.NewBlock(file.path, p.detectedProjectPath, fileBlock, nil, nil, nil),
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

	if p.remoteVariableLoaders != nil {
		for _, loader := range p.remoteVariableLoaders {
			remoteVars, err := loader.Load(RemoteVarLoaderOptions{
				Blocks:      blocks,
				ModulePath:  p.RelativePath(),
				Environment: p.EnvName(),
			})
			if err != nil {
				p.logger.Debug().Msgf("could not load vars from Terraform Cloud: %s", err)
				return combinedVars, err
			}

			for k, v := range remoteVars {
				combinedVars[k] = v
			}
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

	// add a common "env" name to the input vars for the project if it is not
	// explicitly defined. This is done as often users have a common project variable
	// called "env" that is used to define the environment name. Adding this default
	// allows us to have a somewhat sane fallback for situations where we
	// fail to provide the input from a var file or similar config.
	if _, ok := combinedVars["env"]; !ok {
		env := p.workspaceName
		if env == "" {
			env = p.EnvName()
		}

		combinedVars["env"] = cty.StringVal(env)
	}
	if _, ok := combinedVars["environment"]; !ok {
		env := p.workspaceName
		if env == "" {
			env = p.EnvName()
		}

		combinedVars["environment"] = cty.StringVal(env)
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

	opentofuOverrides := make(map[string]struct{})

	for _, info := range fileInfos {
		if info.IsDir() {
			continue
		}

		if modules.HasOpenTofuExtension(info.Name()) {
			name := strings.TrimSuffix(info.Name(), ".json")
			name = strings.TrimSuffix(name, ".tofu")
			opentofuOverrides[name] = struct{}{}
		}
	}

	for _, info := range fileInfos {
		if info.IsDir() {
			continue
		}

		var parseFunc func(filename string) (*hcl.File, hcl.Diagnostics)

		switch {
		case strings.HasSuffix(info.Name(), ".tf"):
			// skip tf file if there's an opentofu override
			if _, ok := opentofuOverrides[strings.TrimSuffix(info.Name(), ".tf")]; ok {
				continue
			}
			parseFunc = hclParser.ParseHCLFile

		case strings.HasSuffix(info.Name(), ".tf.json"):
			// skip tf file if there's an opentofu override
			if _, ok := opentofuOverrides[strings.TrimSuffix(info.Name(), ".tf.json")]; ok {
				continue
			}
			parseFunc = hclParser.ParseJSONFile
		case strings.HasSuffix(info.Name(), ".tofu"):
			parseFunc = hclParser.ParseHCLFile
		case strings.HasSuffix(info.Name(), ".tofu.json"):
			parseFunc = hclParser.ParseJSONFile
		default:
			// this is not a file we can parse:
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
