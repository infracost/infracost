package terraform

import (
	"strings"
	"sync"

	"github.com/infracost/infracost/internal/schema"

	"github.com/infracost/infracost/internal/providers/terraform/aws"
	"github.com/infracost/infracost/internal/providers/terraform/azure"
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

		for _, registryItem := range azure.ResourceRegistry {
			resourceRegistryMap[registryItem.Name] = registryItem
		}
		for _, registryItem := range createFreeResources(azure.FreeResources) {
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

func GetUsageOnlyResources() []string {
	r := []string{}
	r = append(r, aws.UsageOnlyResources...)
	r = append(r, azure.UsageOnlyResources...)
	r = append(r, google.UsageOnlyResources...)
	return r
}

func HasSupportedProvider(rType string) bool {
	return strings.HasPrefix(rType, "aws_") || strings.HasPrefix(rType, "google_") || strings.HasPrefix(rType, "azurerm_")
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
