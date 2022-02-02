package hcl

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	log "github.com/sirupsen/logrus"
)

type Module struct {
	Blocks     Blocks
	RootPath   string
	ModulePath string
}

type ModuleDefinition struct {
	Name       string
	Path       string
	Definition *Block
	Modules    []*Module
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
