package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getComputeSnapshotRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_compute_snapshot",
		RFunc:           newComputeSnapshot,
		ReferenceAttributes: []string{"sourceDisk"},
	}
}

func newComputeSnapshot(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("region").String()

	size := computeSnapshotDiskSize(d)

	r := &google.ComputeSnapshot{
		Address:  d.Address,
		Region:   region,
		DiskSize: size,
	}
	return r
}
