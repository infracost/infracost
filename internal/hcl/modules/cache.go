package modules

import (
	"errors"
	"fmt"
	"sync"

	goversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/rs/zerolog"
)

// Cache is a cache of modules that can be used to lookup modules to check if they've already been loaded.
//
// This only works with modules that have the same identifier. It doesn't cache modules that are used
// multiple times with different identifiers. That is done separately by the PackageFetcher and only
// caches per-run of Infracost, so if you add the same module to your Terraform code it will redownload that module.
// We could optimize it by moving the package fetching cache logic into here, but it would be inconsistent
// with how terraform init works.
type Cache struct {
	keyMap sync.Map
	disco  *Disco
	logger zerolog.Logger
}

// NewCache creates a new cache from a module manifest
func NewCache(disco *Disco, logger zerolog.Logger) *Cache {
	return &Cache{
		disco:  disco,
		logger: logger,
	}
}

func (c *Cache) loadFromManifest(manifest *Manifest) {
	if manifest == nil {
		return
	}

	for _, module := range manifest.Modules {
		c.keyMap.Store(module.Key, module)
	}
}

// lookupModule looks up a module in the cache by its key and checks that the
// source and version are compatible with the module in the cache.
func (c *Cache) lookupModule(key string, moduleCall *tfconfig.ModuleCall) (*ManifestModule, error) {
	value, ok := c.keyMap.Load(key)
	if !ok {
		return nil, errors.New("not in cache")
	}
	manifestModule, _ := value.(*ManifestModule)

	// If the module could be a valid registry module, we should generate the normalized registry source and check against that
	// so we can check the cache against that since we convert to this format before storing in the manifest
	// We don't care about errors here since we only want to check against the registry source if the address is a valid registry address
	var registrySource = ""
	moduleAddr, submodulePath, err := splitModuleSubDir(moduleCall.Source)
	if err == nil {
		registryModuleAddr, err := normalizeRegistrySource(moduleAddr)
		if err == nil {
			registrySource = joinModuleSubDir(registryModuleAddr, submodulePath)
		}
	}

	if manifestModule.Source == moduleCall.Source {
		return checkVersion(moduleCall, manifestModule)
	}

	if manifestModule.Source == registrySource {
		return checkVersion(moduleCall, manifestModule)
	}

	url, _, err := c.disco.ModuleLocation(moduleCall.Source)
	if err != nil {
		c.logger.Debug().Err(err).Msgf("could not fetch module location from source. Proceeding as if source has changed.")
	}

	if manifestModule.Source == url.Location {
		return checkVersion(moduleCall, manifestModule)
	}

	return nil, errors.New("source has changed")
}

func checkVersion(moduleCall *tfconfig.ModuleCall, manifestModule *ManifestModule) (*ManifestModule, error) {
	if moduleCall.Version != "" && manifestModule.Version != "" {
		constraints, err := goversion.NewConstraint(moduleCall.Version)
		if err != nil {
			return nil, fmt.Errorf("invalid version constraint: %w", err)
		}

		version, err := goversion.NewVersion(manifestModule.Version)
		if err != nil {
			return nil, fmt.Errorf("invalid version: %w", err)
		}

		if !constraints.Check(version) {
			return nil, errors.New("version constraint doesn't match")
		}
	}

	return manifestModule, nil
}
