package google

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetContainerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_container_registry",
		RFunc:               NewContainerRegistry,
		ReferenceAttributes: []string{},
	}
}

func NewContainerRegistry(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	components := []*schema.CostComponent{
		dataStorage(d, u),
	}

	components = append(components, operations(d, u)...)
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: components,
		SubResources: []*schema.Resource{
			networkEgress(d, u),
		},
	}
}
