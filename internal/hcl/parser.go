package hcl

import (
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
)

type Option func(p *Parser)

func OptionDoNotSearchTfFiles() Option {
	return func(p *Parser) {
		p.stopOnFirstTf = false
	}
}

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

// Parser is a tool for parsing terraform templates at a given file system location
type Parser struct {
	initialPath    string
	tfvarsPaths    []string
	stopOnFirstTf  bool
	stopOnHCLError bool
	workspaceName  string
}

// New creates a new Parser
func New(initialPath string, options ...Option) *Parser {
	p := &Parser{
		initialPath:   initialPath,
		stopOnFirstTf: true,
		workspaceName: "default",
	}

	for _, option := range options {
		option(p)
	}

	return p
}

func (parser *Parser) parseDirectoryFiles(files []*hcl.File) (Blocks, error) {
	var blocks Blocks

	for _, file := range files {
		fileBlocks, err := LoadBlocksFromFile(file)
		if err != nil {
			if parser.stopOnHCLError {
				return nil, err
			}
			_, _ = fmt.Fprintf(os.Stderr, "WARNING: HCL error: %s\n", err)
			continue
		}
		if len(fileBlocks) > 0 {
			log.Debugf("Added %d blocks from %s...", len(fileBlocks), fileBlocks[0].DefRange.Filename)
		}
		for _, fileBlock := range fileBlocks {
			blocks = append(blocks, NewHCLBlock(fileBlock, nil, nil))
		}
	}

	return blocks, nil
}

// ParseDirectory parses all terraform files within a given directory
func (parser *Parser) ParseDirectory() ([]*Module, error) {
	var blocks Blocks

	log.Debugf("Beginning parse for directory '%s'...", parser.initialPath)
	files, err := LoadDirectory(parser.initialPath, parser.stopOnHCLError)
	if err != nil {
		return nil, err
	}

	parsedBlocks, err := parser.parseDirectoryFiles(files)
	if err != nil {
		return nil, err
	}

	blocks = append(blocks, parsedBlocks...)
	if len(blocks) == 0 && parser.stopOnFirstTf {
		return nil, nil
	}

	log.Debug("Loading TFVars...")

	inputVars, err := loadVars(parser.tfvarsPaths)
	if err != nil {
		return nil, err
	}

	var modulesMetadata *ModulesMetadata
	modulesMetadata, err = loadModuleMetadata(parser.initialPath)
	if err != nil {
		return nil, fmt.Errorf("Error loading module metadata this is required for Infracost to function, %w", err)
	}

	log.Debug("Evaluating expressions...")
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("Error could not evaluate current working directory %w", err)
	}

	evaluator := NewEvaluator(parser.initialPath, parser.initialPath, workingDir, blocks, inputVars, modulesMetadata, nil, parser.stopOnHCLError, parser.workspaceName)
	modules, err := evaluator.EvaluateAll()
	if err != nil {
		return nil, err
	}

	return modules, nil
}

var knownFiles = make(map[string]struct{})

func LoadDirectory(fullPath string, stopOnHCLError bool) ([]*hcl.File, error) {
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
			_, _ = fmt.Fprintf(os.Stderr, "WARNING: HCL error: %s\n", diag)
			continue
		}

		knownFiles[path] = struct{}{}
	}

	var files []*hcl.File
	for _, file := range hclParser.Files() {
		files = append(files, file)
	}

	return files, nil
}

func loadVars(filenames []string) (map[string]cty.Value, error) {
	combinedVars := make(map[string]cty.Value)

	for _, filename := range filenames {
		vars, err := loadTFVars(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to load the tfvars. %s", err.Error())
		}
		for k, v := range vars {
			combinedVars[k] = v
		}
	}

	return combinedVars, nil
}

func loadTFVars(filename string) (map[string]cty.Value, error) {
	inputVars := make(map[string]cty.Value)

	if filename == "" {
		return inputVars, nil
	}

	log.Debugf("loading tfvars-file [%s]", filename)
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	variableFile, _ := hclsyntax.ParseConfig(src, filename, hcl.Pos{Line: 1, Column: 1})
	attrs, _ := variableFile.Body.JustAttributes()

	for _, attr := range attrs {
		log.Debugf("Setting '%s' from tfvars file at %s", attr.Name, filename)
		inputVars[attr.Name], _ = attr.Expr.Value(&hcl.EvalContext{})
	}

	return inputVars, nil
}
