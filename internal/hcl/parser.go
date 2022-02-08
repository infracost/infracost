package hcl

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	log "github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/hcl/modules"
)

type Option func(p *Parser)

func OptionWithTFVarsPaths(paths []string) Option {
	return func(p *Parser) {
		p.tfvarsPaths = paths
	}
}

func OptionStopOnHCLError() Option {
	return func(p *Parser) {
		p.stopOnHCLError = true
	}
}

func OptionWithWorkspaceName(workspaceName string) Option {
	return func(p *Parser) {
		p.workspaceName = workspaceName
	}
}

func TfVarsToOption(varsPaths ...string) (Option, error) {
	for _, p := range varsPaths {
		tfvp, _ := filepath.Abs(p)
		_, err := os.Stat(tfvp)
		if err != nil {
			return nil, fmt.Errorf("passed tfvar file does not exist")
		}
	}

	return OptionWithTFVarsPaths(varsPaths), nil
}

// Parser is a tool for parsing terraform templates at a given file system location.
type Parser struct {
	initialPath    string
	tfvarsPaths    []string
	stopOnHCLError bool
	workspaceName  string
	moduleLoader   *modules.ModuleLoader
}

// New creates a new Parser with the provided options, it inits the workspace as under the default name
// this can be changed using Option.
func New(initialPath string, options ...Option) *Parser {
	p := &Parser{
		initialPath:   initialPath,
		workspaceName: "default",
		moduleLoader:  modules.NewModuleLoader(initialPath),
	}

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
func (parser *Parser) ParseDirectory() ([]*Module, error) {
	log.Debugf("Beginning parse for directory '%s'...", parser.initialPath)

	// load the initial root directory into a list of hcl files
	// at this point these files have no schema associated with them.
	files, err := loadDirectory(parser.initialPath, parser.stopOnHCLError)
	if err != nil {
		return nil, err
	}

	// load the files into given hcl block types. These are then wrapped with *Block structs.
	blocks, err := parser.parseDirectoryFiles(files)
	if err != nil {
		return nil, err
	}

	if len(blocks) == 0 {
		return nil, errors.New("Error no valid tf files found in given path")
	}

	log.Debug("Loading TFVars...")
	inputVars, err := loadVars(parser.tfvarsPaths)
	if err != nil {
		return nil, err
	}

	// load the modules. This downloads any remote modules to the local file system
	modulesManifest, err := parser.moduleLoader.Load()
	if err != nil {
		if !config.IsTest() {
			log.Warnf("Error loading modules. This is required for Infracost to get accurate results: %s", err)
		}
	}

	log.Debug("Evaluating expressions...")
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("Error could not evaluate current working directory %w", err)
	}

	// load an Evaluator with the top level Blocks to begin Context propagation.
	evaluator := NewEvaluator(
		parser.initialPath,
		parser.initialPath,
		workingDir,
		blocks,
		inputVars,
		modulesManifest,
		nil,
		parser.workspaceName,
	)

	modules, err := evaluator.Run()
	if err != nil {
		return nil, err
	}

	return modules, nil
}

func (parser *Parser) parseDirectoryFiles(files []*hcl.File) (Blocks, error) {
	var blocks Blocks

	for _, file := range files {
		fileBlocks, err := loadBlocksFromFile(file)
		if err != nil {
			if parser.stopOnHCLError {
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

func loadVars(filenames []string) (map[string]cty.Value, error) {
	combinedVars := make(map[string]cty.Value)

	for _, filename := range filenames {
		vars, err := loadVarFile(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to load the tfvars. %s", err.Error())
		}

		for k, v := range vars {
			combinedVars[k] = v
		}
	}

	return combinedVars, nil
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
