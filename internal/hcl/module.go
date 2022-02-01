package hcl

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	log "github.com/sirupsen/logrus"
)

type Module struct {
	Blocks     Blocks
	RootPath   string
	ModulePath string
}

func (c *Module) getBlocksByType(blockType string, label string) Blocks {
	var results Blocks

	for _, block := range c.Blocks {
		if block.Type() == blockType && len(block.Labels()) > 0 && block.TypeLabel() == label {
			results = append(results, block)
		}
	}

	return results
}

func (c *Module) getModuleBlocks() Blocks {
	var results Blocks

	for _, block := range c.Blocks {
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

	for _, block := range c.Blocks {
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
		for _, block := range c.Blocks {
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

func getModuleBlocks(b *Block, modulePath string, blocks *Blocks, stopOnHCLError bool) error {
	moduleFiles, err := loadDirectory(modulePath, stopOnHCLError)
	if err != nil {
		return fmt.Errorf("failed to load module %s: %w", b.Label(), err)
	}

	moduleCtx := NewContext(&hcl.EvalContext{}, nil)
	for _, file := range moduleFiles {
		fileBlocks, err := loadBlocksFromFile(file)
		if err != nil {
			if stopOnHCLError {
				return err
			}
			log.Warnf("hcl error loading blocks %s", err)
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
		return nil, fmt.Errorf("metadata file does not exist %w", err)
	}

	f, err := os.Open(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("could not read metadata file %w", err)
	}
	defer func() { _ = f.Close() }()

	var metadata ModulesMetadata
	if err := json.NewDecoder(f).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("could not decode metadata file %w", err)
	}

	return &metadata, nil
}
