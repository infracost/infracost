package hcl

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"

	"github.com/infracost/infracost/internal/hcl/modules"
)

// This sets a global logger for this package, which is a bit of a hack. In the future we should use a context for this.
var log = logrus.StandardLogger().WithField("parser", "terraform_hcl")

type Option func(p *Parser)

// OptionWithTFVarsPaths takes a slice of paths and sets them on the parser relative
// to the Parser initialPath. Paths that don't exist will be ignored.
func OptionWithTFVarsPaths(paths []string) Option {
	return func(p *Parser) {
		var relative []string

		for _, name := range paths {
			tfvp := path.Join(p.initialPath, name)
			_, err := os.Stat(tfvp)
			if err != nil {
				log.Warnf("passed tfvar file does not exist at %s", tfvp)
				continue
			}

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

// OptionWithInputVars takes a cmd line var input values and converts them to cty.Value
// It then sets these as the Parser starting inputVars which be used at the root module evaluation.
func OptionWithInputVars(vs []string) Option {
	return func(p *Parser) {
		ctyVars := make(map[string]cty.Value)
		for _, v := range vs {
			pieces := strings.Split(v, "=")
			if len(pieces) != 2 {
				continue
			}

			ctyVars[pieces[0]] = cty.StringVal(pieces[1])
		}

		p.inputVars = ctyVars
	}
}

func OptionWithWorkspaceName(workspaceName string) Option {
	return func(p *Parser) {
		p.workspaceName = workspaceName
	}
}

// Parser is a tool for parsing terraform templates at a given file system location.
type Parser struct {
	initialPath     string
	defaultVarFiles []string
	tfvarsPaths     []string
	inputVars       map[string]cty.Value
	stopOnHCLError  bool
	workspaceName   string
	moduleLoader    *modules.ModuleLoader
}

// New creates a new Parser with the provided options, it inits the workspace as under the default name
// this can be changed using Option.
func New(initialPath string, options ...Option) *Parser {
	p := &Parser{
		initialPath:   initialPath,
		workspaceName: "default",
		moduleLoader:  modules.NewModuleLoader(initialPath),
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

	return p
}

// ParseDirectory parses all the terraform files in the initalPath into Blocks and then passes them to an Evaluator
// to fill these Blocks with additional Context information. Parser does not parse any blocks outside the root Module.
// It instead leaves ModuleLoader to fetch these Modules on demand. See ModuleLoader.Load for more information.
//
// ParseDirectory returns a list of Module that represent the Terraform Config tree.
func (p *Parser) ParseDirectory() ([]*Module, error) {
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
		return nil, errors.New("Error no valid tf files found in given path")
	}

	log.Debug("Loading TFVars...")
	inputVars, err := p.loadVars(p.tfvarsPaths)
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
		p.initialPath,
		p.initialPath,
		workingDir,
		blocks,
		inputVars,
		modulesManifest,
		nil,
		p.workspaceName,
	)

	modules, err := evaluator.Run()
	if err != nil {
		return nil, err
	}

	return modules, nil
}

func (p *Parser) parseDirectoryFiles(files []*hcl.File) (Blocks, error) {
	var blocks Blocks

	for _, file := range files {
		fileBlocks, err := loadBlocksFromFile(file)
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
				NewHCLBlock(fileBlock, nil, nil),
			)
		}
	}

	return blocks, nil
}

func (p *Parser) loadVars(filenames []string) (map[string]cty.Value, error) {
	combinedVars := make(map[string]cty.Value)

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
		return fmt.Errorf("failed to load the tfvars. %s", err.Error())
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

	log.Debugf("loading tfvars-file [%s]", filename)
	src, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s %w", filename, err)
	}

	variableFile, _ := hclsyntax.ParseConfig(src, filename, hcl.Pos{Line: 1, Column: 1})
	attrs, _ := variableFile.Body.JustAttributes()

	for _, attr := range attrs {
		log.Debugf("Setting '%s' from tfvars file at %s", attr.Name, filename)
		inputVars[attr.Name], _ = attr.Expr.Value(&hcl.EvalContext{})
	}

	return inputVars, nil
}

func loadDirectory(fullPath string, stopOnHCLError bool) ([]*hcl.File, error) {
	hclParser := hclparse.NewParser()

	fileInfos, err := ioutil.ReadDir(fullPath)
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

	files := make([]*hcl.File, 0, len(hclParser.Files()))
	for _, file := range hclParser.Files() {
		files = append(files, file)
	}

	return files, nil
}
