package modules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Manifest is a struct that represents the JSON found in the manifest.json file in the .infracost dir
// It is used for caching the modules that have already been downloaded.
// It uses the same format as the .terraform/modules/modules.json file
type Manifest struct {
	Modules []*ManifestModule `json:"Modules"`
}

// ManifestModule represents a single module in the manifest.json file
type ManifestModule struct {
	Key     string `json:"Key"`
	Source  string `json:"Source"`
	Version string `json:"Version,omitempty"`
	Dir     string `json:"Dir"`
}

// readManifest reads the manifest file from the given path
func readManifest(path string) (*Manifest, error) {
	var manifest Manifest

	data, err := os.ReadFile(path)
	if err != nil {
		return &manifest, fmt.Errorf("Failed to read module manifest: %w", err)
	}

	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return &manifest, fmt.Errorf("Failed to unmarshal module manifest: %w", err)
	}

	return &manifest, err
}

// writeManifest writes the manifest file to the given path
func writeManifest(manifest *Manifest, path string) error {
	b, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("Failed to marshal manifest: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return fmt.Errorf("Failed to create directories for manifest: %w", err)
	}

	err = os.WriteFile(path, b, 0644) // nolint:gosec
	if err != nil {
		return fmt.Errorf("Failed to write manifest: %w", err)
	}

	return nil
}
