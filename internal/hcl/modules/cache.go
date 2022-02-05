package modules

import (
	goversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/pkg/errors"
)

// Cache is a cache of modules that can be used to lookup modules to check
// if they've already been loaded
type Cache struct {
	keyMap map[string]*ManifestModule
}

// NewCacheFromManifest creates a new cache from a module manifest
func NewCacheFromManifest(manifest *Manifest) *Cache {
	cache := &Cache{
		keyMap: make(map[string]*ManifestModule),
	}

	for _, module := range manifest.Modules {
		cache.keyMap[module.Key] = module
	}

	return cache
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
