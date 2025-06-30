package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getComputeImageRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_compute_image",
		RFunc:           newComputeImage,
		ReferenceAttributes: []string{"sourceDisk", "sourceImage", "sourceSnapshot"},
	}
}

func newComputeImage(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("region").String()

	storageSize := computeImageDiskSize(d)

	r := &google.ComputeImage{
		Address:     d.Address,
		Region:      region,
		StorageSize: storageSize,
	}
	return r
}
