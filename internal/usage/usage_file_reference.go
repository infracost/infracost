package usage

import (
	"strings"

	"github.com/infracost/infracost"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
)

type UsageFileReference struct { // nolint:revive
	usageFile *UsageFile
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
		usageFile: usageFile,
	}

	referenceFile.stripResourceNames()
	referenceFile.setDefaultValues()

	return referenceFile, nil
}

func (u *UsageFileReference) stripResourceNames() {
	for _, resourceUsage := range u.usageFile.ResourceUsages {
		resourceType := strings.Split(resourceUsage.Name, ".")[0]
		resourceUsage.Name = resourceType
	}
}

func (u *UsageFileReference) setDefaultValues() {
	for _, resourceUsage := range u.usageFile.ResourceUsages {
		for _, item := range resourceUsage.Items {
			setUsageItemDefaultValues(item)
		}
	}
}

func (u *UsageFileReference) findMatchingResourceUsage(name string) *ResourceUsage {
	addrParts := strings.Split(name, ".")
	if len(addrParts) < 2 {
		return nil
	}

	resourceType := addrParts[len(addrParts)-2]

	for _, resourceUsage := range u.usageFile.ResourceUsages {
		if resourceUsage.Name == resourceType {
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
