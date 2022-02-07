package modules

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
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
		return &manifest, errors.Wrap(err, "Failed to read module manifest")
	}

	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return &manifest, errors.Wrap(err, "Failed to unmarshal module manifest")
	}

	return &manifest, err
}

// writeManifest writes the manifest file to the given path
func writeManifest(manifest *Manifest, path string) error {
	b, err := json.Marshal(manifest)
	if err != nil {
		return errors.Wrap(err, "Failed to marshal manifest")
	}

	err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "Failed to create directories for manifest")
	}

	err = ioutil.WriteFile(path, b, 0644) // nolint:gosec
	if err != nil {
		return errors.Wrap(err, "Failed to write manifest")
	}

	return nil
}
