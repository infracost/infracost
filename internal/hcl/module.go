package hcl

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	log "github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type Module struct {
	blocks     Blocks
	rootPath   string
	modulePath string
}

func NewHCLModule(rootPath string, modulePath string, blocks Blocks) *Module {
	return &Module{
		blocks:     blocks,
		rootPath:   rootPath,
		modulePath: modulePath,
	}
}

func (c *Module) GetBlocks() Blocks {
	return c.blocks
}

func (c *Module) getBlocksByType(blockType string, label string) Blocks {
	var results Blocks

	for _, block := range c.blocks {
		if block.Type() == blockType && len(block.Labels()) > 0 && block.TypeLabel() == label {
			results = append(results, block)
		}
	}

	return results
}

func (c *Module) getModuleBlocks() Blocks {
	var results Blocks

	for _, block := range c.blocks {
		if block.Type() == "module" {
			results = append(results, block)
		}
	}

	return results
}

func (c *Module) GetResourcesByType(label string) Blocks {
	return c.getBlocksByType("resource", label)
}

func (c *Module) GetDatasByType(label string) Blocks {
	return c.getBlocksByType("data", label)
}

func (c *Module) GetProviderBlocksByProvider(providerName string, alias string) Blocks {
	var results Blocks

	for _, block := range c.blocks {
		if block.Type() == "provider" && len(block.Labels()) > 0 && block.TypeLabel() == providerName {
			if alias != "" {
				name := strings.ReplaceAll(alias, providerName+".", "")
				if block.HasChild("alias") && block.GetAttribute("alias").Equals(name) {
					results = append(results, block)
				}
			} else if block.MissingChild("alias") {
				results = append(results, block)
			}
		}
	}

	return results
}

func (c *Module) GetReferencedBlock(referringAttr *Attribute) (*Block, error) {
	for _, ref := range referringAttr.AllReferences() {
		for _, block := range c.blocks {
			if ref.RefersTo(block) {
				return block, nil
			}
		}
	}

	return nil, fmt.Errorf("no referenced block found in '%s'", referringAttr.Name())
}

func (c *Module) GetReferencingResources(originalBlock *Block, referencingLabel string, referencingAttributeName string) (Blocks, error) {
	return c.getReferencingBlocks(originalBlock, "resource", referencingLabel, referencingAttributeName)
}

func (c *Module) GetsModulesBySource(moduleSource string) (Blocks, error) {
	var results Blocks

	modules := c.getModuleBlocks()
	for _, module := range modules {
		if module.HasChild("source") && module.GetAttribute("source").Equals(moduleSource) {
			results = append(results, module)
		}
	}

	return results, nil
}

func (c *Module) getReferencingBlocks(originalBlock *Block, referencingType string, referencingLabel string, referencingAttributeName string) (Blocks, error) {
	blocks := c.getBlocksByType(referencingType, referencingLabel)
	var results Blocks

	for _, block := range blocks {
		attr := block.GetAttribute(referencingAttributeName)
		if attr == nil {
			continue
		}

		if attr.ReferencesBlock(originalBlock) {
			results = append(results, block)
			continue
		}

		for _, ref := range attr.AllReferences() {
			if ref.TypeLabel() == "each" {
				fe := block.GetAttribute("for_each")
				if fe.ReferencesBlock(originalBlock) {
					results = append(results, block)
				}
			}
		}
	}

	return results, nil
}

type ModuleDefinition struct {
	Name       string
	Path       string
	Definition *Block
	Modules    []*Module
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
func (e *Evaluator) loadModule(b *Block, stopOnHCLError bool) (*ModuleDefinition, error) {
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

	var blocks Blocks
	err := getModuleBlocks(b, modulePath, &blocks, stopOnHCLError)
	if err != nil {
		return nil, err
	}
	log.Debugf("Loaded module '%s' (requested at %s)", modulePath, b.Range())

	return &ModuleDefinition{
		Name:       b.Label(),
		Path:       modulePath,
		Definition: b,
		Modules:    []*Module{NewHCLModule(e.projectRootPath, modulePath, blocks)},
	}, nil
}

func getModuleBlocks(b *Block, modulePath string, blocks *Blocks, stopOnHCLError bool) error {
	moduleFiles, err := LoadDirectory(modulePath, stopOnHCLError)
	if err != nil {
		return fmt.Errorf("failed to load module %s: %w", b.Label(), err)
	}

	moduleCtx := NewContext(&hcl.EvalContext{}, nil)
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
			*blocks = append(*blocks, NewHCLBlock(fileBlock, moduleCtx, b))
		}
	}
	return nil
}

type ModulesMetadata struct {
	Modules []ModuleMetadata `json:"Modules"`
}

type ModuleMetadata struct {
	Key     string `json:"Key"`
	Source  string `json:"Source"`
	Version string `json:"Version"`
	Dir     string `json:"Dir"`
}

func loadModuleMetadata(fullPath string) (*ModulesMetadata, error) {
	metadataPath := filepath.Join(fullPath, ".terraform/modules/modules.json")
	if _, err := os.Stat(metadataPath); err != nil {
		return nil, err
	}

	f, err := os.Open(metadataPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var metadata ModulesMetadata
	if err := json.NewDecoder(f).Decode(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}
