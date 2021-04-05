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

func NewComputeImage(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	description := "Storage Image"

	var storageSize *decimal.Decimal
	if computeImageDiskSize(d) != nil {
		storageSize = computeImageDiskSize(d)
	} else if u != nil && u.Get("storage_gb").Exists() {
		storageSize = decimalPtr(decimal.NewFromInt(u.Get("storage_gb").Int()))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: storageImage(region, description, storageSize),
	}
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

func storageImage(region string, description string, storageSize *decimal.Decimal) []*schema.CostComponent {
	return []*schema.CostComponent{
		{
			Name:            "Storage",
			Unit:            "GB-months",
			UnitMultiplier:  1,
			MonthlyQuantity: storageSize,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr(region),
				Service:       strPtr("Compute Engine"),
				ProductFamily: strPtr("Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", Value: strPtr(description)},
				},
			},
		},
	}
}
