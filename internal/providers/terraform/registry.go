package terraform

import (
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
	})

	return &resourceRegistryMap
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
	return func(d *schema.ResourceData) []string {
		return []string{d.Get("id").String()}
	}
}

// CreatePartialResource creates a partial resource based on the given
// ResourceData and UsageData. If the resource type is found in the registry, it
// checks if it has NoPrice flag set. If so, it creates a PartialResource with
// NoPrice set to true. Otherwise, it checks if it has a CoreRFunc (preferred way
// to create provider-agnostic resources) and uses it to generate a CoreResource
// if possible. If CoreResource is not available, it uses the RFunc (legacy way)
// to create a regular Resource. If UsageData is provided, it calculates the
// estimation summary and adds it to the created resource. Finally, it returns
// the created PartialResource object.
func (r *ResourceRegistryMap) CreatePartialResource(d *schema.ResourceData, u *schema.UsageData) *schema.PartialResource {
	if registryItem, ok := (*r)[d.Type]; ok {
		if registryItem.NoPrice {
			return &schema.PartialResource{
				ResourceData: d,
				Resource: &schema.Resource{
					Name:        d.Address,
					IsSkipped:   true,
					NoPrice:     true,
					SkipMessage: "Free resource.",
					Metadata:    d.Metadata,
				},
				CloudResourceIDs: registryItem.CloudResourceIDFunc(d),
			}
		}

		// Use the CoreRFunc to generate a CoreResource if possible.  This is
		// the new/preferred way to create provider-agnostic resources that
		// support advanced features such as Infracost Cloud usage estimates
		// and actual costs.
		if registryItem.CoreRFunc != nil {
			coreRes := registryItem.CoreRFunc(d)
			if coreRes != nil {
				return &schema.PartialResource{ResourceData: d, CoreResource: coreRes, CloudResourceIDs: registryItem.CloudResourceIDFunc(d)}
			}
		} else {
			res := registryItem.RFunc(d, u)
			if res != nil {
				if u != nil {
					res.EstimationSummary = u.CalcEstimationSummary()
				}

				return &schema.PartialResource{ResourceData: d, Resource: res, CloudResourceIDs: registryItem.CloudResourceIDFunc(d)}
			}
		}
	}

	return &schema.PartialResource{
		ResourceData: d,
		Resource: &schema.Resource{
			Name:        d.Address,
			IsSkipped:   true,
			SkipMessage: "This resource is not currently supported",
			Metadata:    d.Metadata,
		},
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
