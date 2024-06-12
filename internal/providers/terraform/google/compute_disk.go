package google

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getComputeDiskRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_compute_disk",
		CoreRFunc:           newComputeDisk,
		ReferenceAttributes: []string{"image", "snapshot"},
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			region := d.Get("region").String()

			zone := d.Get("zone").String()
			if zone != "" {
				region = zoneToRegion(zone)
			}

			return region
		},
	}
}

func newComputeDisk(d *schema.ResourceData) schema.CoreResource {
	diskType := d.Get("type").String()
	size := computeDiskSize(d)

	iops := computeIOPS(d, diskType, size)

	r := &google.ComputeDisk{
		Address: d.Address,
		Region:  d.Region,
		Type:    diskType,
		Size:    size,
		IOPS:    iops,
	}
	return r
}

func computeDiskSize(d *schema.ResourceData) float64 {
	if d.Get("size").Exists() {
		return d.Get("size").Float()
	}

	if len(d.References("image")) > 0 {
		return computeImageDiskSize(d.References("image")[0])
	}

	if len(d.References("snapshot")) > 0 {
		return computeSnapshotDiskSize(d.References("snapshot")[0])
	}

	return defaultDiskSize(d.Get("type").String())
}

func defaultDiskSize(diskType string) float64 {
	diskType = strings.ToLower(diskType)
	if diskType == "pd-balanced" || diskType == "pd-ssd" {
		return 100
	}

	if diskType == "pd-extreme" || diskType == "hyperdisk-extreme" {
		return 1000
	}

	// if diskType is not specificed, default value is pd-standard
	return 500
}

func computeImageDiskSize(d *schema.ResourceData) float64 {
	if d.Get("disk_size_gb").Exists() {
		return d.Get("disk_size_gb").Float()
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

	return 0
}

func computeSnapshotDiskSize(d *schema.ResourceData) float64 {
	if len(d.References("source_disk")) > 0 {
		return computeDiskSize(d.References("source_disk")[0])
	}

	return 0
}

func computeIOPS(d *schema.ResourceData, diskType string, diskSize float64) int64 {
	if diskType == "pd-extreme" || diskType == "hyperdisk-extreme" {

		if d.Get("provisioned_iops").Exists() {
			return d.Get("provisioned_iops").Int()
		}

		return defaultIOPS(diskType, diskSize)
	}

	return 0
}

func defaultIOPS(diskType string, diskSize float64) int64 {
	if diskType == "pd-extreme" {
		return 100000
	}

	if diskType == "hyperdisk-extreme" {

		if iops := diskSize * 1200; iops < 350000 {
			return int64(iops)
		}

		return 350000
	}

	return 0
}
