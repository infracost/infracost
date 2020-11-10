package google

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetComputeSnapshotRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_compute_snapshot",
		RFunc:               NewComputeSnapshot,
		ReferenceAttributes: []string{"source_disk"},
	}
}

func NewComputeSnapshot(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	// Not yet implemented, but added here so that the references can be used for google_compute_disk
	return nil
}

func computeSnapshotDiskSize(d *schema.ResourceData) *decimal.Decimal {
	if len(d.References("source_disk")) > 0 {
		return computeDiskSize(d.References("source_disk")[0])
	}

	return nil
}
