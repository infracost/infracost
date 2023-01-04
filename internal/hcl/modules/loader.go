package modules

import (
	"crypto/md5" //nolint
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	getter "github.com/hashicorp/go-getter"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	intSync "github.com/infracost/infracost/internal/sync"
	"github.com/infracost/infracost/internal/ui"
)

var (
	// downloadDir is name of the directory where remote modules are download
	downloadDir = ".infracost/terraform_modules"
	// manifestPath is the name of the module manifest file which stores the metadata of the modules
	manifestPath = ".infracost/terraform_modules/manifest.json"
	// tfManifestPath is the name of the terraform module manifest file which stores the metadata of the modules
	tfManifestPath = ".terraform/modules/modules.json"

	supportedManifestVersion = "2.0"
)

// ModuleLoader handles the loading of Terraform modules. It supports local, registry and other remote modules.
//
// The path should be the root directory of the Terraform project. We use a distinct module loader per Terraform project,
// because at the moment the cache is per project. The cache reads the manifest.json file from the path's
// .infracost/terraform_modules directory. We could implement a global cache in the future, but for now have decided
// to go with the same approach as Terraform.
type ModuleLoader struct {
	NewSpinner ui.SpinnerFunc

	// cachePath is the path to the directory that Infracost will download modules to.
	// This is normally the top level directory of a multi-project environment, where the
	// Infracost config file resides or project auto-detection starts from.
	cachePath string
	cache     *Cache
	sync      *intSync.KeyMutex

	packageFetcher *PackageFetcher
	registryLoader *RegistryLoader
	logger         *logrus.Entry
}

// NewModuleLoader constructs a new module loader
func NewModuleLoader(cachePath string, credentialsSource *CredentialsSource, logger *logrus.Entry, moduleSync *intSync.KeyMutex) *ModuleLoader {
	fetcher := NewPackageFetcher(logger)
	// we need to have a disco for each project that has defined credentials
	d := NewDisco(credentialsSource, logger)

	m := &ModuleLoader{
		cachePath:      cachePath,
		cache:          NewCache(d, logger),
		packageFetcher: fetcher,
		logger:         logger,
		sync:           moduleSync,
	}

	m.registryLoader = NewRegistryLoader(fetcher, d, logger)

	return m
}

// downloadDir returns the path to the directory where remote modules are downloaded relative to the current working directory
func (m *ModuleLoader) downloadDir() string {
	return filepath.Join(m.cachePath, downloadDir)
}

// manifestFilePath is the path to the module manifest file relative to the current working directory
func (m *ModuleLoader) manifestFilePath(projectPath string) string {
	if m.cachePath == projectPath {
		return filepath.Join(m.cachePath, manifestPath)
	}

	rel, _ := filepath.Rel(m.cachePath, projectPath)
	sum := md5.Sum([]byte(rel)) //nolint
	return filepath.Join(m.cachePath, ".infracost/terraform_modules/", fmt.Sprintf("manifest-%x.json", sum))
}

// tfManifestFilePath is the path to the Terraform module manifest file relative to the current working directory.
func (m *ModuleLoader) tfManifestFilePath(path string) string {
	return filepath.Join(path, tfManifestPath)
}

// Load loads the modules from the given path.
// For each module it checks if the module has already been downloaded, by checking if iut exists in the manifest
// If not then it downloads the module from the registry or from a remote source and updates the module manifest with the latest metadata.
func (m *ModuleLoader) Load(path string) (man *Manifest, err error) {
	defer func() {
		if man != nil {
			man.cachePath = m.cachePath
		}
	}()

	if m.NewSpinner != nil {
		spin := m.NewSpinner("Downloading Terraform modules")
		defer spin.Success()
	}

	manifest := &Manifest{}
	manifestFilePath := m.manifestFilePath(path)
	_, err = os.Stat(manifestFilePath)
	if errors.Is(err, os.ErrNotExist) {
		m.logger.Debug("No existing module manifest file found")

		tfManifestFilePath := m.tfManifestFilePath(path)
		_, err = os.Stat(tfManifestFilePath)
		if err == nil {
			manifest, err = readManifest(tfManifestFilePath)
			if err == nil {
				// let's make the module dirs relative to the path directory as later
				// we'll look up the modules based on the cache path at the Infracost root (where the infracost.yml
				// resides or where the --path autodetect started for multi-project)
				for i, module := range manifest.Modules {
					dir := path
					if m.cachePath != "" {
						dir, _ = filepath.Rel(m.cachePath, path)
					}

					manifest.Modules[i].Dir = filepath.Join(dir, module.Dir)
				}

				return manifest, nil
			}

			m.logger.WithError(err).Debug("error reading terraform module manifest")
		}
	} else if err != nil {
		m.logger.WithError(err).Debug("error checking for existing module manifest")
	} else {
		manifest, err = readManifest(manifestFilePath)
		if err != nil {
			m.logger.WithError(err).Debug("could not read module manifest")
		}

		if manifest.Version != supportedManifestVersion {
			manifest = &Manifest{cachePath: m.cachePath}
		}
	}
	m.cache.loadFromManifest(manifest)

	metadatas, err := m.loadModules(path, "")
	if err != nil {
		return nil, err
	}

	manifest.Modules = metadatas
	manifest.Path = path
	manifest.Version = supportedManifestVersion

	err = writeManifest(manifest, manifestFilePath)
	if err != nil {
		m.logger.WithError(err).Debug("error writing module manifest")
	}

	return manifest, nil
}

// loadModules recursively loads the modules from the given path.
func (m *ModuleLoader) loadModules(path string, prefix string) ([]*ManifestModule, error) {
	manifestModules := make([]*ManifestModule, 0)

	module, diags := tfconfig.LoadModule(path)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to inspect module path %s diag: %w", path, diags.Err())
	}

	numJobs := len(module.ModuleCalls)
	jobs := make(chan *tfconfig.ModuleCall, numJobs)
	for _, moduleCall := range module.ModuleCalls {
		jobs <- moduleCall
	}
	close(jobs)

	errGroup := &errgroup.Group{}
	manifestMu := sync.Mutex{}

	for i := 0; i < getProcessCount(); i++ {
		errGroup.Go(func() error {
			for moduleCall := range jobs {
				metadata, err := m.loadModule(moduleCall, path, prefix)
				if err != nil {
					return err
				}

				manifestMu.Lock()
				manifestModules = append(manifestModules, metadata)
				manifestMu.Unlock()

				moduleDir := filepath.Join(m.cachePath, metadata.Dir)
				nestedManifestModules, err := m.loadModules(moduleDir, metadata.Key+".")
				if err != nil {
					return err
				}

				manifestMu.Lock()
				manifestModules = append(manifestModules, nestedManifestModules...)
				manifestMu.Unlock()
			}

			return nil
		})
	}

	err := errGroup.Wait()
	if err != nil {
		return manifestModules, fmt.Errorf("could not load modules for path %s %w", path, err)
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
	source := moduleCall.Source

	manifestModule, err := m.cache.lookupModule(key, moduleCall)
	if err == nil {
		m.logger.Debugf("module %s already loaded", key)

		// Test if we can actually load the module. If not, then we should try re-loading it.
		// This can happen if the directory the module was downloaded to has been deleted and moved
		// so the existing manifest.json is out-of-date.
		_, diags := tfconfig.LoadModule(path.Join(m.cachePath, manifestModule.Dir))
		if !diags.HasErrors() {
			return manifestModule, err
		}

		m.logger.Debugf("module %s cannot be loaded, re-loading: %s", key, diags.Err())
	} else {
		m.logger.Debugf("module %s needs loading: %s", key, err.Error())
	}

	manifestModule = &ManifestModule{
		Key:    key,
		Source: source,
	}

	if m.isLocalModule(moduleCall) {
		dir, err := filepath.Rel(m.cachePath, filepath.Join(parentPath, source))
		if err != nil {
			return nil, err
		}

		m.logger.Debugf("loading local module %s from %s", key, dir)
		manifestModule.Dir = path.Clean(dir)
		return manifestModule, nil
	}

	moduleAddr, submodulePath, err := splitModuleSubDir(source)
	if err != nil {
		return nil, err
	}

	hash := fmt.Sprintf("%x", md5.Sum([]byte(moduleAddr+moduleCall.Version))) //nolint
	dest := filepath.Join(m.downloadDir(), hash)

	// lock the module address so that we don't interact with an incomplete download.
	unlock := m.sync.Lock(moduleAddr)
	defer unlock()

	moduleDownloadDir, err := filepath.Rel(m.cachePath, dest)
	if err != nil {
		return nil, err
	}
	manifestModule.Dir = path.Clean(filepath.Join(moduleDownloadDir, submodulePath))

	lookupResult, err := m.registryLoader.lookupModule(moduleAddr, moduleCall.Version)
	if err != nil {
		return nil, fmt.Errorf("error looking up registry module %s: %w", key, err)
	}

	if lookupResult.OK {
		// The moduleCall.Source might not have the registry hostname if it is using the default registry
		// so we set the source here to the lookup result's source which always includes the registry hostname.
		manifestModule.Source = joinModuleSubDir(lookupResult.ModuleURL.RawSource, submodulePath)
		manifestModule.Version = lookupResult.Version

		_, err = os.Stat(dest)
		if err == nil {
			return manifestModule, nil
		}

		err = m.registryLoader.downloadModule(lookupResult, dest)
		if err != nil {
			return nil, fmt.Errorf("failed to download registry module %s: %w", key, err)
		}

		return manifestModule, nil
	}

	m.logger.Debugf("Detected %s as remote module", key)
	m.logger.Debugf("Downloading module %s from remote %s", key, source)

	_, err = os.Stat(dest)
	if err == nil {
		return manifestModule, nil
	}

	err = m.packageFetcher.fetch(moduleAddr, dest)
	if err != nil {
		return nil, fmt.Errorf("failed to download remote module %s: %w", key, err)
	}

	return manifestModule, nil
}

// isLocalModule checks if the module is a local module by checking
// if the module source starts with any known local prefixes
func (m *ModuleLoader) isLocalModule(moduleCall *tfconfig.ModuleCall) bool {
	return strings.HasPrefix(moduleCall.Source, "./") ||
		strings.HasPrefix(moduleCall.Source, "../") ||
		strings.HasPrefix(moduleCall.Source, ".\\") ||
		strings.HasPrefix(moduleCall.Source, "..\\")
}

func splitModuleSubDir(moduleSource string) (string, string, error) {
	moduleAddr, submodulePath := getter.SourceDirSubdir(moduleSource)
	if strings.HasPrefix(submodulePath, "../") {
		return "", "", fmt.Errorf("invalid submodule path '%s'", submodulePath)
	}

	return moduleAddr, submodulePath, nil
}

func joinModuleSubDir(moduleAddr string, submodulePath string) string {
	if submodulePath != "" {
		return fmt.Sprintf("%s//%s", moduleAddr, submodulePath)
	}

	return moduleAddr
}

func getProcessCount() int {
	numWorkers := 4
	numCPU := runtime.NumCPU()

	if numCPU*4 > numWorkers {
		numWorkers = numCPU * 4
	}
	if numWorkers > 16 {
		numWorkers = 16
	}

	return numWorkers
}
