package usage

import (
	"context"
	"sort"

	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type SyncResult struct {
	ResourceCount    int
	EstimationCount  int
	EstimationErrors map[string]error
}

func SyncUsageData(usageFile *UsageFile, projects []*schema.Project) (*SyncResult, error) {
	referenceFile, err := LoadReferenceFile()
	if err != nil {
		return nil, err
	}
	referenceFile.SetDefaultValues()

	// TODO: update this when we properly support multiple projects in usage
	resources := make([]*schema.Resource, 0)
	for _, project := range projects {
		resources = append(resources, project.Resources...)
	}

	syncResult := syncResourceUsages(usageFile, resources, referenceFile)

	return syncResult, nil
}

func syncResourceUsages(usageFile *UsageFile, resources []*schema.Resource, referenceFile *ReferenceFile) *SyncResult {
	syncResult := &SyncResult{
		EstimationErrors: make(map[string]error),
	}

	existingResourceUsagesMap := resourceUsagesMap(usageFile.ResourceUsages)
	resourcesUsages := make([]*ResourceUsage, 0, len(resources))

	existingResourceOrder := make([]string, 0, len(usageFile.ResourceUsages))
	for _, resourceUsage := range usageFile.ResourceUsages {
		existingResourceOrder = append(existingResourceOrder, resourceUsage.Name)
	}

	sortResourcesExistingFirst(resources, existingResourceOrder)

	for _, resource := range resources {
		resourceUsage := &ResourceUsage{
			Name: resource.Name,
		}

		// Merge the usage schema from the reference usage file
		refResourceUsage := referenceFile.FindMatchingResourceUsage(resource.Name)
		if refResourceUsage != nil {
			mergeResourceUsages(resourceUsage, refResourceUsage)
		}

		// Merge the usage schema from the resource struct
		mergeResourceUsages(resourceUsage, &ResourceUsage{
			Name:  resource.Name,
			Items: resource.UsageSchema,
		})

		// Merge any existing resource usage
		existingResourceUsage := existingResourceUsagesMap[resource.Name]
		if existingResourceUsage != nil {
			mergeResourceUsages(resourceUsage, existingResourceUsage)
		}

		syncResult.ResourceCount++
		if resource.EstimateUsage != nil {
			syncResult.EstimationCount++

			resourceUsageMap := resourceUsage.Map()
			err := resource.EstimateUsage(context.TODO(), resourceUsageMap)
			if err != nil {
				syncResult.EstimationErrors[resource.Name] = err
				log.Warnf("Error estimating usage for resource %s: %v", resource.Name, err)
			}

			// Merge in the estimated usage
			estimatedUsageData := schema.NewUsageData(resource.Name, schema.ParseAttributes(resourceUsageMap))
			mergeResourceUsageWithUsageData(resourceUsage, estimatedUsageData)
		}

		resourcesUsages = append(resourcesUsages, resourceUsage)
	}

	usageFile.ResourceUsages = resourcesUsages

	return syncResult
}

func mergeResourceUsages(dest *ResourceUsage, src *ResourceUsage) {
	if dest == nil || src == nil {
		return
	}

	destItemMap := make(map[string]*schema.UsageItem, len(dest.Items))
	for _, item := range dest.Items {
		destItemMap[item.Key] = item
	}

	for _, srcItem := range src.Items {
		destItem, ok := destItemMap[srcItem.Key]
		if !ok {
			destItem = &schema.UsageItem{Key: srcItem.Key}
			dest.Items = append(dest.Items, destItem)
		}

		destItem.ValueType = srcItem.ValueType

		if srcItem.Description != "" {
			destItem.Description = srcItem.Description
		}

		if srcItem.ValueType == schema.SubResourceUsage {
			if srcItem.DefaultValue != nil {
				srcDefaultValue := srcItem.DefaultValue.(*ResourceUsage)
				if destItem.DefaultValue == nil {
					destItem.DefaultValue = &ResourceUsage{
						Name: srcDefaultValue.Name,
					}
				}
				mergeResourceUsages(destItem.DefaultValue.(*ResourceUsage), srcDefaultValue)
			}

			if srcItem.Value != nil {
				srcValue := srcItem.Value.(*ResourceUsage)
				if destItem.Value == nil {
					destItem.Value = destItem.DefaultValue
				}
				if destItem.Value == nil {
					destItem.Value = &ResourceUsage{
						Name: srcValue.Name,
					}
				}
				mergeResourceUsages(destItem.Value.(*ResourceUsage), srcValue)
			}
		} else {
			if srcItem.DefaultValue != nil {
				destItem.DefaultValue = srcItem.DefaultValue
			}

			if srcItem.Value != nil {
				destItem.Value = srcItem.Value
			}
		}
	}
}

func mergeResourceUsageWithUsageData(resourceUsage *ResourceUsage, usageData *schema.UsageData) {
	if usageData == nil {
		return
	}

	for _, item := range resourceUsage.Items {
		var val interface{}

		switch item.ValueType {
		case schema.Int64:
			if v := usageData.GetInt(item.Key); v != nil {
				val = *v
			}
		case schema.Float64:
			if v := usageData.GetFloat(item.Key); v != nil {
				val = *v
			}
		case schema.String:
			if v := usageData.GetString(item.Key); v != nil {
				val = *v
			}
		case schema.StringArray:
			if v := usageData.GetStringArray(item.Key); v != nil {
				val = *v
			}
		case schema.SubResourceUsage:
			subUsageMap := usageData.Get(item.Key).Map()
			subExisting := schema.NewUsageData(item.Key, subUsageMap)

			subResourceUsage := &ResourceUsage{
				Name: item.Key,
			}
			// If the item has a value, use it as the base
			if item.Value != nil {
				subResourceUsage = item.Value.(*ResourceUsage)
			}

			// If the item has no value but does have a default value, then merge in
			// any keys that exist in the existing usage that match a key in the default value
			if item.Value == nil && item.DefaultValue != nil {
				for _, subItem := range item.DefaultValue.(*ResourceUsage).Items {
					if subExisting.Get(subItem.Key).Type != gjson.Null {
						subResourceUsage.Items = append(subResourceUsage.Items, subItem)
					}
				}
			}

			mergeResourceUsageWithUsageData(subResourceUsage, subExisting)
			val = item.Value
		}

		item.Value = val
	}
}

// sortResourcesExistingFirst sorts the resources by the existing order first, and then the rest by name
func sortResourcesExistingFirst(resources []*schema.Resource, existingResourceOrder []string) {
	sort.Slice(resources, func(i, j int) bool {
		existingIndexI := indexOf(resources[i].Name, existingResourceOrder)
		existingIndexJ := indexOf(resources[j].Name, existingResourceOrder)

		// If both resources are in the existing resource order, sort by the existing resource order
		if existingIndexI != -1 && existingIndexJ != -1 {
			return existingIndexI < existingIndexJ
		}

		// If one resource is in the existing resource order, sort it first
		if existingIndexI == -1 && existingIndexJ != -1 {
			return false
		}

		if existingIndexJ == -1 && existingIndexI != -1 {
			return true
		}

		// Otherwise sort by name
		return resources[i].Name < resources[j].Name
	})
}

func indexOf(s string, arr []string) int {
	for k, v := range arr {
		if s == v {
			return k
		}
	}
	return -1
}
