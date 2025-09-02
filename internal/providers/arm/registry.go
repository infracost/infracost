package arm

import (
	"github.com/infracost/infracost/internal/providers/arm/azure"
	"github.com/infracost/infracost/internal/schema"
)

type RegistryItemMap map[string]*schema.RegistryItem

var (
	ResourceRegistryMap = buildResourceRegistryMap()
)

func buildResourceRegistryMap() *RegistryItemMap {
	resourceRegistryMap := make(RegistryItemMap)

	for _, registryItem := range azure.ResourceRegistry {
		if registryItem.CloudResourceIDFunc == nil {
			registryItem.CloudResourceIDFunc = azure.DefaultCloudResourceIDFunc
		}
		resourceRegistryMap[registryItem.Name] = registryItem
		resourceRegistryMap[registryItem.Name].DefaultRefIDFunc = azure.GetDefaultRefIDFunc
	}
	for _, registryItem := range createFreeResources(azure.FreeResources, azure.GetDefaultRefIDFunc, azure.DefaultCloudResourceIDFunc) {
		resourceRegistryMap[registryItem.Name] = registryItem
	}

	return &resourceRegistryMap
}

// GetRegion returns the region lookup function for the given resource data type if it exists.
func (r *RegistryItemMap) GetRegion(resourceDataType string) schema.RegionLookupFunc {
	item, ok := (*r)[resourceDataType]
	if ok {
		return item.GetRegion
	}

	return nil
}

func createFreeResources(l []string, defaultRefsFunc schema.ReferenceIDFunc, resourceIdFunc schema.CloudResourceIDFunc) []*schema.RegistryItem {
	freeResources := make([]*schema.RegistryItem, 0)
	for _, resourceName := range l {
		freeResources = append(freeResources, &schema.RegistryItem{
			Name:                resourceName,
			NoPrice:             true,
			Notes:               []string{"Free resource."},
			DefaultRefIDFunc:    defaultRefsFunc,
			CloudResourceIDFunc: resourceIdFunc,
		})
	}
	return freeResources
}
