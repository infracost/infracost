package pulumi

import (
	"strings"
	"sync"

	"github.com/infracost/infracost/internal/schema"

	"github.com/infracost/infracost/internal/providers/pulumi/aws"
	"github.com/infracost/infracost/internal/providers/pulumi/types"
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
			resourceRegistryMap[registryItem.Name].DefaultRefIDFunc = GetDefaultRefIDFunc
		}
		for _, registryItem := range createFreeResources(aws.FreeResources, GetDefaultRefIDFunc) {
			resourceRegistryMap[registryItem.Name] = registryItem
		}
		for _, registryItem := range createFreeResources(types.PulumiFreeResources, GetDefaultRefIDFunc) {
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

func HasSupportedProvider(rType string) bool {
	return strings.HasPrefix(rType, "aws_") ||
		strings.HasPrefix(rType, "google_") ||
		strings.HasPrefix(rType, "azurerm_") ||
		strings.HasPrefix(rType, "pulumi_") ||
		strings.HasPrefix(rType, "kubernetes_")
}

func createFreeResources(l []string, defaultRefsFunc schema.ReferenceIDFunc) []*schema.RegistryItem {
	freeResources := make([]*schema.RegistryItem, 0)
	for _, resourceName := range l {
		freeResources = append(freeResources, &schema.RegistryItem{
			Name:             resourceName,
			NoPrice:          true,
			Notes:            []string{"Free resource."},
			DefaultRefIDFunc: defaultRefsFunc,
		})
	}
	return freeResources
}
func GetDefaultRefIDFunc(d *schema.ResourceData) []string {
	defaultRefs := []string{d.RawValues.Get("urn").String()}
	return defaultRefs
}

func (r *ResourceRegistryMap) GetReferenceAttributes(resourceDataType string) []string {
	var refAttrs []string
	item, ok := (*r)[resourceDataType]
	if ok {
		refAttrs = item.ReferenceAttributes
	}
	return refAttrs
}

func (r *ResourceRegistryMap) GetCustomRefIDFunc(resourceDataType string) schema.ReferenceIDFunc {
	item, ok := (*r)[resourceDataType]
	if ok {
		return item.CustomRefIDFunc
	}
	return nil
}

func (r *ResourceRegistryMap) GetDefaultRefIDFunc(resourceDataType string) schema.ReferenceIDFunc {
	item, ok := (*r)[resourceDataType]
	if ok {
		return item.DefaultRefIDFunc
	}
	return nil
}
