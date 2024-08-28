package crossplane

import (
	"sync"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/providers/crossplane/azure"
	"github.com/infracost/infracost/internal/schema"
)

type ResourceRegistryMap map[string]*schema.RegistryItem

var (
	resourceRegistryMap ResourceRegistryMap
	once                sync.Once
)

func GetResourceRegistryMap() *ResourceRegistryMap {
	once.Do(func() {
		resourceRegistryMap = make(ResourceRegistryMap)

		// Merge all resource registries

		for _, registryItem := range azure.ResourceRegistry {
			logging.Logger.Debug().Msgf("Registering resource: %s", registryItem.Name)
			resourceRegistryMap[registryItem.Name] = registryItem
		}
		for _, registryItem := range createFreeResources(azure.FreeResources) {
			logging.Logger.Debug().Msgf("Registering free resource: %s", registryItem.Name)
			resourceRegistryMap[registryItem.Name] = registryItem
		}

	})

	return &resourceRegistryMap
}

func GetUsageOnlyResources() []string {
	r := []string{}
	r = append(r, azure.UsageOnlyResources...)
	return r
}

// func HasSupportedProvider(rType string) bool {
// 	return strings.Contains(rType, ".azure.crossplane.io")
// }

func createFreeResources(l []string) []*schema.RegistryItem {
	freeResources := make([]*schema.RegistryItem, 0)
	for _, resourceName := range l {
		freeResources = append(freeResources, &schema.RegistryItem{
			Name:    resourceName,
			NoPrice: true,
			Notes:   []string{"Free resource."},
		})
		logging.Logger.Debug().Msgf("Creating free resource entry: %s", resourceName)
	}
	return freeResources
}
