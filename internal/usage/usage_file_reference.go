package usage

import (
	"strings"

	"github.com/infracost/infracost"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
)

type UsageFileReference struct { // nolint:revive
	*UsageFile
}

func LoadReferenceFile() (*UsageFileReference, error) {
	contents := infracost.GetReferenceUsageFileContents()
	if contents == nil {
		return &UsageFileReference{}, errors.New("Could not find reference usage file")
	}

	usageFile, err := LoadUsageFileFromString(string(*contents))
	if err != nil {
		return &UsageFileReference{}, err
	}

	referenceFile := &UsageFileReference{
		UsageFile: usageFile,
	}

	return referenceFile, nil
}

// SetDefaultValues updates the reference file to strip the values and set the default values
func (u *UsageFileReference) SetDefaultValues() {
	for _, resourceUsage := range u.ResourceUsages {
		for _, item := range resourceUsage.Items {
			setUsageItemDefaultValues(item)
		}
	}
}

// FindMatchingResourceUsage returns the matching resource usage for the given resource name
// by looking for a resource with the same resource type
func (u *UsageFileReference) FindMatchingResourceUsage(name string) *ResourceUsage {
	addrParts := strings.Split(name, ".")
	if len(addrParts) < 2 {
		return nil
	}

	wantResourceType := addrParts[len(addrParts)-2]

	for _, resourceUsage := range u.ResourceUsages {
		resourceType := strings.Split(resourceUsage.Name, ".")[0]
		if resourceType == wantResourceType {
			return resourceUsage
		}
	}
	return nil
}

func setUsageItemDefaultValues(item *schema.UsageItem) {
	if item == nil {
		return
	}

	switch item.ValueType {
	case schema.Float64:
		item.DefaultValue = 0.0
	case schema.Int64:
		item.DefaultValue = 0
	case schema.Items:
		if item.Value != nil {
			for _, subItem := range item.Value.([]*schema.UsageItem) {
				setUsageItemDefaultValues(subItem)
			}
		}

		item.DefaultValue = item.Value
	default:
		item.DefaultValue = item.Value
	}

	item.Value = nil
}
