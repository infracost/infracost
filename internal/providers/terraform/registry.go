package terraform

import (
	"strings"
	"sync"

	"github.com/infracost/infracost/internal/schema"

	"github.com/infracost/infracost/internal/providers/terraform/aws"
	"github.com/infracost/infracost/internal/providers/terraform/google"
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

		for _, registryItem := range google.ResourceRegistry {
			resourceRegistryMap[registryItem.Name] = registryItem
		}
		for _, registryItem := range createFreeResources(google.FreeResources) {
			resourceRegistryMap[registryItem.Name] = registryItem
		}
	})

	return &resourceRegistryMap
}

func HasSupportedProvider(rType string) bool {
	return strings.HasPrefix(rType, "aws_") || strings.HasPrefix(rType, "google_")
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
