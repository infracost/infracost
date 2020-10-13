package terraform

import (
	"sync"

	"github.com/infracost/infracost/internal/schema"

	"github.com/infracost/infracost/internal/providers/terraform/aws"
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

		// AWS
		for _, registryItem := range aws.ResourceRegistry {
			resourceRegistryMap[registryItem.Name] = registryItem
		}
		for _, registryItem := range aws.GetFreeResources() {
			resourceRegistryMap[registryItem.Name] = registryItem
		}

	})

	return &resourceRegistryMap
}
