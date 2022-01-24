package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	log "github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"

	"github.com/infracost/infracost/internal/hcl/block"
)

type ModuleDefinition struct {
	Name       string
	Path       string
	Definition block.Block
	Modules    []block.Module
}

// LoadModules reads all module blocks and loads the underlying modules, adding blocks to e.moduleBlocks
func (e *Evaluator) loadModules(stopOnHCLError bool) []*ModuleDefinition {

	blocks := e.blocks

	var moduleDefinitions []*ModuleDefinition

	expanded := e.expandBlocks(blocks.OfType("module"))

	for _, moduleBlock := range expanded {
		if moduleBlock.Label() == "" {
			continue
		}
		moduleDefinition, err := e.loadModule(moduleBlock, stopOnHCLError)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "WARNING: Failed to load module: %s\n", err)
			continue
		}
		moduleDefinitions = append(moduleDefinitions, moduleDefinition)
	}

	return moduleDefinitions
}

// takes in a module "x" {} block and loads resources etc. into e.moduleBlocks - additionally returns variables to add to ["module.x.*"] variables
func (e *Evaluator) loadModule(b block.Block, stopOnHCLError bool) (*ModuleDefinition, error) {

	if b.Label() == "" {
		return nil, fmt.Errorf("module without label at %s", b.Range())
	}

	var source string
	attrs := b.Attributes()
	for _, attr := range attrs {
		if attr.Name() == "source" {
			sourceVal := attr.Value()
			if sourceVal.Type() == cty.String {
				source = sourceVal.AsString()
			}
		}
	}

	if source == "" {
		return nil, fmt.Errorf("could not read module source attribute at %s", b.Range().String())
	}

	var modulePath string

	if e.moduleMetadata != nil {
		// if we have module metadata we can parse all the modules as they'll be cached locally!
		for _, module := range e.moduleMetadata.Modules {
			reg := "registry.terraform.io/" + source
			if module.Source == source || module.Source == reg {
				modulePath = filepath.Clean(filepath.Join(e.projectRootPath, module.Dir))
				break
			}
		}
	}

	if modulePath == "" {
		// if we have no metadata, we can only support modules available on the local filesystem
		// users wanting this feature should run a `terraform init` before running infracost to cache all modules locally
		if !strings.HasPrefix(source, fmt.Sprintf(".%c", os.PathSeparator)) && !strings.HasPrefix(source, fmt.Sprintf("..%c", os.PathSeparator)) {
			reg := "registry.terraform.io/" + source
			return nil, fmt.Errorf("missing module with source '%s %s' -  try to 'terraform init' first", reg, source)
		}

		// combine the current calling module with relative source of the module
		modulePath = filepath.Join(e.modulePath, source)
	}

	var blocks block.Blocks
	err := getModuleBlocks(b, modulePath, &blocks, stopOnHCLError)
	if err != nil {
		return nil, err
	}
	log.Debugf("Loaded module '%s' (requested at %s)", modulePath, b.Range())

	return &ModuleDefinition{
		Name:       b.Label(),
		Path:       modulePath,
		Definition: b,
		Modules:    []block.Module{block.NewHCLModule(e.projectRootPath, modulePath, blocks)},
	}, nil
}

func getModuleBlocks(b block.Block, modulePath string, blocks *block.Blocks, stopOnHCLError bool) error {
	moduleFiles, err := LoadDirectory(modulePath, stopOnHCLError)
	if err != nil {
		return fmt.Errorf("failed to load module %s: %w", b.Label(), err)
	}

	moduleCtx := block.NewContext(&hcl.EvalContext{}, nil)
	for _, file := range moduleFiles {
		fileBlocks, err := LoadBlocksFromFile(file)
		if err != nil {
			if stopOnHCLError {
				return err
			}
			_, _ = fmt.Fprintf(os.Stderr, "WARNING: HCL error: %s\n", err)
			continue
		}
		if len(fileBlocks) > 0 {
			log.Debugf("Added %d blocks from %s...", len(fileBlocks), fileBlocks[0].DefRange.Filename)
		}
		for _, fileBlock := range fileBlocks {
			*blocks = append(*blocks, block.NewHCLBlock(fileBlock, moduleCtx, b))
		}
	}
	return nil
}
