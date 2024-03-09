package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getCloudRunServiceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_cloud_run_service",
		CoreRFunc: newCloudRunService,
	}
}

func newCloudRunService(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("region").String()
	return &google.CloudRunService{
		Address: d.Address,
		Region:  region,
	}
}
