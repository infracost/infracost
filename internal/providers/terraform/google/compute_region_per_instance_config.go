package google

import (
	"github.com/infracost/infracost/internal/schema"
)

func getComputeRegionPerInstanceConfigRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_compute_region_per_instance_config",
		NoPrice:             true,
		ReferenceAttributes: []string{"region_instance_group_manager"},
		Notes:               []string{"Free resource."},
	}
}
