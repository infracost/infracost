package modules

import (
	goversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/pkg/errors"
)

// Cache is a cache of modules that can be used to lookup modules to check if they've already been loaded.
//
// This only works with modules that have the same identifier. It doesn't cache modules that are used
// multiple times with different identifiers. That is done separately by the PackageFetcher and only
// caches per-run of Infracost, so if you add the same module to your Terraform code it will redownload that module.
// We could optimize it by moving the package fetching cache logic into here, but it would be inconsistent
// with how terraform init works.
type Cache struct {
	keyMap map[string]*ManifestModule
}

// NewCacheFromManifest creates a new cache from a module manifest
func NewCache() *Cache {
	return &Cache{
		keyMap: make(map[string]*ManifestModule),
	}
}

func (c *Cache) loadFromManifest(manifest *Manifest) {
	if manifest == nil {
		return
	}

	for _, module := range manifest.Modules {
		c.keyMap[module.Key] = module
	}
}

// lookupModule looks up a module in the cache by its key and checks that the
// source and version are compatible with the module in the cache.
func (c *Cache) lookupModule(key string, moduleCall *tfconfig.ModuleCall) (*ManifestModule, error) {
	manifestModule, ok := c.keyMap[key]

	if !ok {
		return nil, errors.New("not in cache")
	}

	if manifestModule.Source != moduleCall.Source {
		return nil, errors.New("source has changed")
	}

	if moduleCall.Version != "" && manifestModule.Version != "" {
		constraints, err := goversion.NewConstraint(moduleCall.Version)
		if err != nil {
			return nil, errors.Wrap(err, "invalid version constraint")
		}

		version, err := goversion.NewVersion(manifestModule.Version)
		if err != nil {
			return nil, errors.Wrap(err, "invalid version")
		}

		if !constraints.Check(version) {
			return nil, errors.New("version constraint doesn't match")
		}
	}

	return manifestModule, nil
}
