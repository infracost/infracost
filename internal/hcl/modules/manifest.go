package modules

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Manifest is a struct that represents the JSON found in the manifest.json file in the .infracost dir
// It is used for caching the modules that have already been downloaded.
// It uses the same format as the .terraform/modules/modules.json file
type Manifest struct {
	cachePath string

	Path    string            `json:"Path"`
	Version string            `json:"Version"`
	Modules []*ManifestModule `json:"Modules"`
}

func (m Manifest) Get(key string) ManifestModule {
	for _, module := range m.Modules {
		if module.Key == key {
			loc := filepath.Clean(filepath.Join(m.cachePath, module.Dir))
			return ManifestModule{
				Key:         module.Key,
				Source:      module.Source,
				Version:     module.Version,
				Dir:         loc,
				DownloadURL: module.DownloadURL,
			}
		}
	}

	return ManifestModule{}
}

// ManifestModule represents a single module in the manifest.json file
type ManifestModule struct {
	Key            string `json:"Key"`
	Source         string `json:"Source"`
	Version        string `json:"Version,omitempty"`
	Dir            string `json:"Dir"`
	DownloadURL    string
	IsSourceMapped bool `json:"-"`
}

func (m ManifestModule) URL() string {
	if IsLocalModule(m.Source) {
		return ""
	}

	remoteSource := m.Source

	if m.DownloadURL != "" {
		remoteSource = m.DownloadURL
	}

	remoteSource, _, _ = splitModuleSubDir(remoteSource)
	remoteSource = strings.TrimPrefix(remoteSource, "git::")
	remoteSource = strings.TrimPrefix(remoteSource, "gcs::")
	remoteSource = strings.TrimPrefix(remoteSource, "s3::")

	u, err := url.Parse(remoteSource)
	if err == nil {
		u.RawQuery = ""
		return u.String()
	}

	return remoteSource
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
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directories for manifest: %w", err)
	}

	err = os.WriteFile(path, b, 0644) // nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}
