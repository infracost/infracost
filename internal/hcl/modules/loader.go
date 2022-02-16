package modules

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	getter "github.com/hashicorp/go-getter"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	log "github.com/sirupsen/logrus"
)

// downloadDir is name of the directory where remote modules are download
var downloadDir = ".infracost/terraform_modules"

// manifestPath is the name of the module manifest file which stores the metadata of the modules
var manifestPath = ".infracost/terraform_modules/manifest.json"

// ModuleLoader handles the loading of Terraform modules. It supports local, registry and other remote modules.
//
// The path should be the root directory of the Terraform project. We use a distinct module loader per Terraform project,
// because at the moment the cache is per project. The cache reads the manifest.json file from the path's
// .infracost/terraform_modules directory. We could implement a global cache in the future, but for now have decided
// to go with the same approach as Terraform.
type ModuleLoader struct {
	Path           string
	cache          *Cache
	packageFetcher *PackageFetcher
	registryLoader *RegistryLoader
}

// NewModuleLoader constructs a new module loader
func NewModuleLoader(path string) *ModuleLoader {
	fetcher := NewPackageFetcher()

	return &ModuleLoader{
		Path:           path,
		cache:          NewCache(),
		packageFetcher: fetcher,
		registryLoader: NewRegistryLoader(fetcher),
	}
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

	m.cache.loadFromManifest(manifest)

	metadatas, err := m.loadModules(m.Path, "")
	if err != nil {
		return nil, err
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

		// Test if we can actually load the module. If not, then we should try re-loading it.
		// This can happen if the directory the module was downloaded to has been deleted and moved
		// so the existing manifest.json is out-of-date.
		_, diags := tfconfig.LoadModule(path.Join(m.Path, manifestModule.Dir))
		if !diags.HasErrors() {
			return manifestModule, err
		}

		log.Debugf("Module %s cannot be loaded, re-loading: %s", key, diags.Err())
	} else {
		log.Debugf("Module %s needs loading: %s", key, err.Error())
	}

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
		manifestModule.Dir = path.Clean(dir)
		return manifestModule, nil
	}

	dest := filepath.Join(m.downloadDir(), key)
	moduleDownloadDir, err := filepath.Rel(m.Path, dest)
	if err != nil {
		return nil, err
	}

	moduleAddr, submodulePath := getter.SourceDirSubdir(moduleCall.Source)
	if strings.HasPrefix(submodulePath, "../") {
		return nil, fmt.Errorf("Invalid submodule path '%s'", submodulePath)
	}

	if submodulePath != "" {
		moduleDownloadDir = filepath.Join(moduleDownloadDir, submodulePath)
	}

	manifestModule.Dir = path.Clean(moduleDownloadDir)

	lookupResult, err := m.registryLoader.lookupModule(moduleAddr, moduleCall.Version)
	if err == nil {
		log.Debugf("Downloading module %s from registry URL %s", key, lookupResult.DownloadURL)
		err = m.registryLoader.downloadModule(lookupResult.DownloadURL, dest)
		if err != nil {
			return nil, err
		}

		// The moduleCall.Source might not have the registry hostname if it is using the default registry
		// so we set the source here to the lookup result's source which always includes the registry hostname.
		manifestModule.Source = lookupResult.Source
		if submodulePath != "" {
			manifestModule.Source = fmt.Sprintf("%s//%s", manifestModule.Source, submodulePath)
		}

		manifestModule.Version = lookupResult.Version
		return manifestModule, nil
	}

	log.Debugf("Module %s not recognized as registry module, treating as remote module: %s", key, err.Error())
	log.Debugf("Downloading module %s from remote %s", key, moduleCall.Source)
	err = m.packageFetcher.fetch(moduleAddr, dest)
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
