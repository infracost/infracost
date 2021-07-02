package cloudformation

import (
	"sync"

	"github.com/infracost/infracost/internal/schema"

	"github.com/infracost/infracost/internal/providers/cloudformation/aws"
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
		for _, registryItem := range aws.ResourceRegistry {
			resourceRegistryMap[registryItem.Name] = registryItem
		}
		for _, registryItem := range createFreeResources(aws.FreeResources) {
			resourceRegistryMap[registryItem.Name] = registryItem
		}
	})

	return &resourceRegistryMap
}

func GetUsageOnlyResources() []string {
	r := []string{}
	r = append(r, aws.UsageOnlyResources...)
	return r
}

func createFreeResources(l []string) []*schema.RegistryItem {
	freeResources := make([]*schema.RegistryItem, 0)
	for _, resourceName := range l {
		freeResources = append(freeResources, &schema.RegistryItem{
			Name:    resourceName,
			NoPrice: true,
			Notes:   []string{"Free resource."},
		})
	}
	return freeResources
}
