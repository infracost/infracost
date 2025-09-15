package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getCloudSchedulerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_cloud_scheduler",
		CoreRFunc: NewCloudScheduler,
	}
}

func NewCloudScheduler(d *schema.ResourceData) schema.CoreResource {
	r := &google.CloudScheduler{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	return r
}
