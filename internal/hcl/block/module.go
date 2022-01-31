package block

import (
	"fmt"
	"strings"
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
				if block.HasChild("alias") && block.GetAttribute("alias").Equals(strings.ReplaceAll(alias, fmt.Sprintf("%s.", providerName), "")) {
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
		} else {
			for _, ref := range attr.AllReferences() {
				if ref.TypeLabel() == "each" {
					fe := block.GetAttribute("for_each")
					if fe.ReferencesBlock(originalBlock) {
						results = append(results, block)
					}
				}
			}
		}
	}
	return results, nil
}
