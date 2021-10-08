package usage

import (
	"context"

	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"
)

type SyncResult struct {
	ResourceCount    int
	EstimationCount  int
	EstimationErrors map[string]error
}

func (u *UsageFile) SyncUsageData(projects []*schema.Project) (*SyncResult, error) {
	referenceFile, err := LoadReferenceFile()
	if err != nil {
		return nil, err
	}

	// TODO: update this when we properly support multiple projects in usage
	resources := make([]*schema.Resource, 0)
	for _, project := range projects {
		resources = append(resources, project.Resources...)
	}

	syncResult := u.syncResourceUsages(resources, referenceFile)

	return syncResult, nil
}

func (u *UsageFile) syncResourceUsages(resources []*schema.Resource, referenceFile *UsageFileReference) *SyncResult {
	syncResult := &SyncResult{
		EstimationErrors: make(map[string]error),
	}

	existingResourceUsagesMap := resourceUsagesMap(u.ResourceUsages)
	resourcesUsages := make([]*ResourceUsage, 0, len(resources))

	for _, resource := range resources {
		resourceUsage := &ResourceUsage{
			Name: resource.Name,
		}

		// Merge the usage schema from the reference usage file
		refResourceUsage := referenceFile.findMatchingResourceUsage(resource.Name)
		if refResourceUsage != nil {
			mergeResourceUsageItems(resourceUsage, refResourceUsage)
		}

		// Merge the usage schema from the resource struct
		mergeResourceUsageItems(resourceUsage, &ResourceUsage{
			Name:  resource.Name,
			Items: resource.UsageSchema,
		})

		// Merge any existing resource usage
		existingResourceUsage := existingResourceUsagesMap[resource.Name]
		if existingResourceUsage != nil {
			mergeResourceUsageItems(resourceUsage, existingResourceUsage)
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

	u.ResourceUsages = resourcesUsages

	return syncResult
}

func mergeResourceUsageItems(dest *ResourceUsage, src *ResourceUsage) {
	if dest == nil || src == nil {
		return
	}

	dest.Items = mergeUsageItems(dest.Items, src.Items)
}

func mergeUsageItems(destItems []*schema.UsageItem, srcItems []*schema.UsageItem) []*schema.UsageItem {
	destItemMap := make(map[string]*schema.UsageItem, len(destItems))
	for _, item := range destItems {
		destItemMap[item.Key] = item
	}

	for _, srcItem := range srcItems {
		destItem, ok := destItemMap[srcItem.Key]
		if !ok {
			destItem = &schema.UsageItem{Key: srcItem.Key}
			destItems = append(destItems, destItem)
		}

		destItem.ValueType = srcItem.ValueType

		if srcItem.ValueType == schema.Items {
			if srcItem.DefaultValue != nil {
				srcDefaultValue := srcItem.DefaultValue.([]*schema.UsageItem)
				if destItem.DefaultValue == nil {
					destItem.DefaultValue = make([]*schema.UsageItem, 0)
				}
				destItem.DefaultValue = mergeUsageItems(destItem.DefaultValue.([]*schema.UsageItem), srcDefaultValue)
			}

			if srcItem.Value != nil {
				srcValue := srcItem.Value.([]*schema.UsageItem)
				if destItem.Value == nil {
					destItem.Value = destItem.DefaultValue
				}
				if destItem.Value == nil {
					destItem.Value = make([]*schema.UsageItem, 0)
				}
				destItem.Value = mergeUsageItems(destItem.Value.([]*schema.UsageItem), srcValue)
			}
		} else {
			destItem.DefaultValue = srcItem.DefaultValue
			destItem.Value = srcItem.Value
		}
	}

	return destItems
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
		case schema.StringArray:
			if v := usageData.GetStringArray(item.Key); v != nil {
				val = *v
			}
		case schema.Items:
			subResourceUsage := &ResourceUsage{}
			subExisting := schema.NewUsageData(item.Key, usageData.Get(item.Key).Map())
			mergeResourceUsageWithUsageData(subResourceUsage, subExisting)
		}

		item.Value = val
	}
}
