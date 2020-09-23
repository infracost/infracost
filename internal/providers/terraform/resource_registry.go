package terraform

import (
	"sync"

	"github.com/infracost/infracost/pkg/schema"

	"github.com/infracost/infracost/internal/providers/terraform/aws"
)

type resourceRegistryMapSingleton map[string]*schema.RegistryItem

var (
	resourceRegistryMap resourceRegistryMapSingleton
	once                sync.Once
)

func getResourceRegistryMap() *resourceRegistryMapSingleton {
	once.Do(func() {
		resourceRegistryMap = make(resourceRegistryMapSingleton)
		// Merge all resource registries

		// AWS
		for _, registryItem := range aws.ResourceRegistry {
			resourceRegistryMap[registryItem.Name] = registryItem
			for _, alias := range registryItem.Aliases {
				resourceRegistryMap[alias] = registryItem
			}
		}

	})

	return &resourceRegistryMap
}
