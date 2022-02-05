package modules

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// downloadDir is name of the directory where remote modules are download
var downloadDir = ".infracost/terraform_modules"

// manifestPath is the name of the module manifest file which stores the metadata of the modules
var manifestPath = ".infracost/terraform_modules/manifest.json"

// ModuleLoader handles the loading of Terraform modules. It supports local, registry and other remote modules.
type ModuleLoader struct {
	Path  string
	cache *Cache
}

// downloadDir returns the path to the directory where remote modules are downloaded relative to the current working directory
func (m *ModuleLoader) downloadDir() string {
	return filepath.Join(m.Path, downloadDir)
}

// manifestFilePath is the path to the module manifest file relative to the current working directory
func (m *ModuleLoader) manifestFilePath() string {
	return filepath.Join(m.Path, manifestPath)
}

// Load loads the modules from the given path.
// For each module it checks if the module has already been downloaded, by checking if iut exists in the manifest
// If not then it downloads the module from the registry or from a remote source and updates the module manifest with the latest metadata.
func (m *ModuleLoader) Load() (*Manifest, error) {
	var manifest *Manifest

	_, err := os.Stat(m.manifestFilePath())
	if errors.Is(err, os.ErrNotExist) {
		log.Debugf("No existing module manifest file found")
	} else if err != nil {
		log.Debugf("Error checking for existing module manifest: %s", err)
	} else {
		manifest, err = readManifest(m.manifestFilePath())
		if err != nil {
			log.Debugf("Error reading module manifest: %s", err)
		}
	}

	if manifest == nil {
		manifest = &Manifest{}
	}

	m.cache = NewCacheFromManifest(manifest)

	metadatas, err := m.loadModules(m.Path, "")
	if err != nil {
		return manifest, nil
	}

	manifest.Modules = metadatas

	err = writeManifest(manifest, m.manifestFilePath())
	if err != nil {
		log.Debugf("Error writing module manifest: %s", err)
	}

	return manifest, nil
}

// loadModules recursively loads the modules from the given path.
func (m *ModuleLoader) loadModules(path string, prefix string) ([]*ManifestModule, error) {
	manifestModules := make([]*ManifestModule, 0)

	module, diags := tfconfig.LoadModule(path)
	if diags.HasErrors() {
		return nil, diags.Err()
	}

	for _, moduleCall := range module.ModuleCalls {
		metadata, err := m.loadModule(moduleCall, path, prefix)
		if err != nil {
			return nil, err
		}

		manifestModules = append(manifestModules, metadata)

		nestedManifestModules, err := m.loadModules(filepath.Join(m.Path, metadata.Dir), metadata.Key+".")
		if err != nil {
			return nil, err
		}

		manifestModules = append(manifestModules, nestedManifestModules...)
	}

	return manifestModules, nil
}

// loadModule loads the module metadata from the given module call.
// It works by doing the following:
// 1. Checks if the module is already downloaded and the version/source has not changed.
// 2. Checks if the module is a local module.
// 3. Checks if the module is a registry module and downloads it.
// 4. Checks if the module is a remote module and downloads it.
func (m *ModuleLoader) loadModule(moduleCall *tfconfig.ModuleCall, parentPath string, prefix string) (*ManifestModule, error) {
	key := prefix + moduleCall.Name

	manifestModule, err := m.cache.lookupModule(key, moduleCall)
	if err == nil {
		log.Debugf("Module %s already loaded", key)
		return manifestModule, err
	}

	log.Debugf("Module %s needs loaded: %s", key, err.Error())

	manifestModule = &ManifestModule{
		Key:    key,
		Source: moduleCall.Source,
	}

	if m.isLocalModule(moduleCall) {
		dir, err := filepath.Rel(m.Path, filepath.Join(parentPath, moduleCall.Source))
		if err != nil {
			return nil, err
		}

		log.Debugf("Loading local module %s from %s", key, dir)
		manifestModule.Dir = dir
		return manifestModule, nil
	}

	dest := filepath.Join(m.downloadDir(), key)
	dir, err := filepath.Rel(m.Path, dest)
	if err != nil {
		return nil, err
	}
	manifestModule.Dir = dir

	registryLoader := NewRegistryLoader(dest)
	lookupResult, err := registryLoader.lookupModule(moduleCall)
	if err != nil {
		log.Debugf("Module %s not recognized as registry module, treating as remote module: %s", key, err.Error())
	} else {
		log.Debugf("Downloading module %s from registry URL %s", key, lookupResult.DownloadURL)
		err = registryLoader.downloadModule(lookupResult.DownloadURL)
		if err != nil {
			return nil, err
		}

		// The moduleCall.Source might not have the registry hostname if it is using the default registry
		// so we set the source here to the lookup result's source which always includes the registry hostname.
		manifestModule.Source = lookupResult.Source
		manifestModule.Version = lookupResult.Version
		return manifestModule, nil
	}

	log.Debugf("Downloading module %s from remote %s", key, moduleCall.Source)
	remoteLoader := NewRemoteLoader(dest)
	err = remoteLoader.downloadModule(moduleCall)
	if err != nil {
		return nil, err
	}

	return manifestModule, nil
}

// isLocalModule checks if the module is a local module by checking
// if the module source starts with any known local prefixes
func (m *ModuleLoader) isLocalModule(moduleCall *tfconfig.ModuleCall) bool {
	return (strings.HasPrefix(moduleCall.Source, "./") ||
		strings.HasPrefix(moduleCall.Source, "../") ||
		strings.HasPrefix(moduleCall.Source, ".\\") ||
		strings.HasPrefix(moduleCall.Source, "..\\"))
}
