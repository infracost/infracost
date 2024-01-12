package idem

import (
	"sync"

	"github.com/infracost/infracost/internal/schema"

	"github.com/infracost/infracost/internal/providers/idem/aws"
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
			if registryItem.CloudResourceIDFunc == nil {
				registryItem.CloudResourceIDFunc = aws.DefaultCloudResourceIDFunc
			}
			resourceRegistryMap[registryItem.Name] = registryItem
			resourceRegistryMap[registryItem.Name].DefaultRefIDFunc = aws.GetDefaultRefIDFunc
		}
		for _, registryItem := range createFreeResources(aws.FreeResources) {
			if registryItem.CloudResourceIDFunc == nil {
				registryItem.CloudResourceIDFunc = aws.DefaultCloudResourceIDFunc
			}
			resourceRegistryMap[registryItem.Name] = registryItem
			resourceRegistryMap[registryItem.Name].DefaultRefIDFunc = aws.GetDefaultRefIDFunc
		}
	})

	return &resourceRegistryMap
}

func (r *ResourceRegistryMap) GetDefaultRefIDFunc(resourceDataType string) schema.ReferenceIDFunc {
	item, ok := (*r)[resourceDataType]
	if ok {
		return item.DefaultRefIDFunc
	}
	return func(d *schema.ResourceData) []string {
		return []string{d.Get("resource_id").String()}
	}
}

func (r *ResourceRegistryMap) GetCustomRefIDFunc(resourceDataType string) schema.ReferenceIDFunc {
	item, ok := (*r)[resourceDataType]
	if ok {
		return item.CustomRefIDFunc
	}
	return nil
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

func (r *ResourceRegistryMap) GetReferenceAttributes(resourceDataType string) []string {
	var refAttrs []string
	item, ok := (*r)[resourceDataType]
	if ok {
		refAttrs = item.ReferenceAttributes
	}
	return refAttrs
}
