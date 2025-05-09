package google

import (
	"github.com/infracost/infracost/internal/schema"
)

func getComputePerInstanceConfigRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_compute_per_instance_config",
		NoPrice:             true,
		ReferenceAttributes: []string{"instanceGroupManager"},
		Notes:               []string{"Free resource."},
	}
}
