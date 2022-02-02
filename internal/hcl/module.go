package hcl

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	log "github.com/sirupsen/logrus"
)

// ModuleCall represents a call to a defined Module by a parent Module.
type ModuleCall struct {
	// Name the name of the module as specified a the point of definition.
	Name string
	// Path is the path to the local directory containing the HCL for the Module.
	Path string
	// Definition is the actual Block where the ModuleCall happens in a hcl.File
	Definition *Block
	// Modules contains the parsed modules that are part of this ModuleCall. This can contain
	// more than one Module as it will also contain a list of the child Modules that have been
	// called within this Module. The Module at position 0 is the root Module.
	Modules []*Module
}

// Module encapsulates all the Blocks that are part of a Module in a Terraform project.
type Module struct {
	Blocks     Blocks
	RootPath   string
	ModulePath string
}

// getModuleBlocks loads all the Blocks for the module at the given path
func (b *Block) getModuleBlocks(modulePath string) (Blocks, error) {
	var blocks Blocks
	moduleFiles, err := loadDirectory(modulePath, true)
	if err != nil {
		return blocks, fmt.Errorf("failed to load module %s: %w", b.Label(), err)
	}

	moduleCtx := NewContext(&hcl.EvalContext{}, nil)
	for _, file := range moduleFiles {
		fileBlocks, err := loadBlocksFromFile(file)
		if err != nil {
			return blocks, err
		}

		if len(fileBlocks) > 0 {
			log.Debugf("Added %d blocks from %s...", len(fileBlocks), fileBlocks[0].DefRange.Filename)
		}

		for _, fileBlock := range fileBlocks {
			blocks = append(blocks, NewHCLBlock(fileBlock, moduleCtx, b))
		}
	}

	return blocks, err
}

// ModulesMetadata is a struct that represents the JSON found in the modules.json file int the .terraform dir.
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
