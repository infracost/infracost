package usage

import (
	"strings"

	"github.com/infracost/infracost"
	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
)

// ReferenceFile represents the reference example usage file
type ReferenceFile struct { // nolint:revive
	*UsageFile
}

// LoadReferenceFile loads the reference example usage file
func LoadReferenceFile() (*ReferenceFile, error) {
	contents := infracost.GetReferenceUsageFileContents()
	if contents == nil {
		return &ReferenceFile{}, errors.New("Could not find reference usage file")
	}

	usageFile, err := LoadUsageFileFromString(string(*contents))
	if err != nil {
		return &ReferenceFile{}, err
	}

	referenceFile := &ReferenceFile{
		UsageFile: usageFile,
	}

	return referenceFile, nil
}

// SetDefaultValues updates the reference file to strip the values and set the default values
func (u *ReferenceFile) SetDefaultValues() {
	for _, resourceUsage := range u.ResourceUsages {
		for _, item := range resourceUsage.Items {
			setUsageItemDefaultValues(item)
		}
	}
	for _, resourceUsage := range u.ResourceTypeUsages {
		for _, item := range resourceUsage.Items {
			setUsageItemDefaultValues(item)
		}
	}
}

// FindMatchingResourceUsage returns the matching resource usage for the given resource name
// by looking for a resource with the same resource type
func (u *ReferenceFile) FindMatchingResourceUsage(name string) *ResourceUsage {
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

func (u *ReferenceFile) FindMatchingResourceTypeUsage(resourceType string) *ResourceUsage {
	for _, resourceUsage := range u.ResourceTypeUsages {
		if resourceUsage.Name == resourceType {
			return resourceUsage
		}
	}
	for _, resourceUsage := range u.ResourceUsages {
		referenceResourceType := strings.Split(resourceUsage.Name, ".")[0]
		if referenceResourceType == resourceType {
			return resourceUsage
		}
	}
	return nil
}

// setUsageItemDefaultValues recursively sets the default values for the given usage item
func setUsageItemDefaultValues(item *schema.UsageItem) {
	if item == nil {
		return
	}

	switch item.ValueType {
	case schema.Float64:
		item.DefaultValue = 0.0
	case schema.Int64:
		item.DefaultValue = 0
	case schema.SubResourceUsage:
		if item.Value != nil {
			for _, subItem := range item.Value.(*ResourceUsage).Items {
				setUsageItemDefaultValues(subItem)
			}
		}

		item.DefaultValue = item.Value
	default:
		item.DefaultValue = item.Value
	}

	item.Value = nil
}
