package google

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetComputeImageRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_compute_image",
		RFunc:               NewComputeImage,
		ReferenceAttributes: []string{"source_disk", "source_image", "source_snapshot"},
	}
}

func NewComputeImage(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	// Not yet implemented, but added here so that the references can be used for google_compute_disk
	return nil
}

func computeImageDiskSize(d *schema.ResourceData) *decimal.Decimal {
	if d.Get("disk_size_gb").Exists() {
		return decimalPtr(decimal.NewFromFloat(d.Get("disk_size_gb").Float()))
	}

	if len(d.References("source_disk")) > 0 {
		return computeDiskSize(d.References("source_disk")[0])
	}

	if len(d.References("source_image")) > 0 {
		return computeImageDiskSize(d.References("source_image")[0])
	}

	if len(d.References("source_snapshot")) > 0 {
		return computeSnapshotDiskSize(d.References("source_snapshot")[0])
	}

	return nil
}
