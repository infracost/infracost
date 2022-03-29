package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getContainerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_container_registry",
		RFunc:               NewContainerRegistry,
		ReferenceAttributes: []string{},
	}
}
func NewContainerRegistry(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.ContainerRegistry{
		Address:      d.Address,
		Region:       d.Get("region").String(),
		Location:     d.Get("location").String(),
		StorageClass: d.Get("storage_class").String(),
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
