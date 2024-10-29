package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getComputeMachineImageRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_compute_machine_image",
		CoreRFunc: newComputeMachineImage,
	}
}

func newComputeMachineImage(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("region").String()

	r := &google.ComputeMachineImage{
		Address: d.Address,
		Region:  region,
	}
	return r
}
