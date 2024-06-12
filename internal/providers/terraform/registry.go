package terraform

import (
	"github.com/infracost/infracost/internal/schema"

	"github.com/infracost/infracost/internal/providers/terraform/aws"
	"github.com/infracost/infracost/internal/providers/terraform/azure"
	"github.com/infracost/infracost/internal/providers/terraform/google"
)

type RegistryItemMap map[string]*schema.RegistryItem

var (
	ResourceRegistryMap = buildResourceRegistryMap()
)

func buildResourceRegistryMap() *RegistryItemMap {
	resourceRegistryMap := make(RegistryItemMap)

	// Merge all resource registries
	for _, registryItem := range aws.ResourceRegistry {
		if registryItem.CloudResourceIDFunc == nil {
			registryItem.CloudResourceIDFunc = aws.DefaultCloudResourceIDFunc
		}
		resourceRegistryMap[registryItem.Name] = registryItem
		resourceRegistryMap[registryItem.Name].DefaultRefIDFunc = aws.GetDefaultRefIDFunc
	}
	for _, registryItem := range createFreeResources(aws.FreeResources, aws.GetDefaultRefIDFunc, aws.DefaultCloudResourceIDFunc) {
		resourceRegistryMap[registryItem.Name] = registryItem
	}

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

	for _, registryItem := range google.ResourceRegistry {
		if registryItem.CloudResourceIDFunc == nil {
			registryItem.CloudResourceIDFunc = google.DefaultCloudResourceIDFunc
		}
		resourceRegistryMap[registryItem.Name] = registryItem
		resourceRegistryMap[registryItem.Name].DefaultRefIDFunc = google.GetDefaultRefIDFunc
	}
	for _, registryItem := range createFreeResources(google.FreeResources, google.GetDefaultRefIDFunc, google.DefaultCloudResourceIDFunc) {
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

func (r *RegistryItemMap) GetReferenceAttributes(resourceDataType string) []string {
	var refAttrs []string
	item, ok := (*r)[resourceDataType]
	if ok {
		refAttrs = item.ReferenceAttributes
	}
	return refAttrs
}

func (r *RegistryItemMap) GetCustomRefIDFunc(resourceDataType string) schema.ReferenceIDFunc {
	item, ok := (*r)[resourceDataType]
	if ok {
		return item.CustomRefIDFunc
	}
	return nil
}

func (r *RegistryItemMap) GetDefaultRefIDFunc(resourceDataType string) schema.ReferenceIDFunc {
	item, ok := (*r)[resourceDataType]
	if ok {
		return item.DefaultRefIDFunc
	}
	return func(d *schema.ResourceData) []string {
		return []string{d.Get("id").String()}
	}
}

func GetUsageOnlyResources() []string {
	r := []string{}
	r = append(r, aws.UsageOnlyResources...)
	r = append(r, azure.UsageOnlyResources...)
	r = append(r, google.UsageOnlyResources...)
	return r
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
