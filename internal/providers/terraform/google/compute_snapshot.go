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

func NewComputeSnapshot(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	description := "Storage PD Snapshot"

	var snapshotDiskSize *decimal.Decimal
	if computeSnapshotDiskSize(d) != nil {
		snapshotDiskSize = computeSnapshotDiskSize(d)
	} else if u != nil && u.Get("storage_gb").Exists() {
		snapshotDiskSize = decimalPtr(decimal.NewFromInt(u.Get("storage_gb").Int()))
	}

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Storage",
				Unit:            "GB-months",
				UnitMultiplier:  1,
				MonthlyQuantity: snapshotDiskSize,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr(region),
					Service:       strPtr("Compute Engine"),
					ProductFamily: strPtr("Storage"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", Value: strPtr(description)},
					},
				},
				PriceFilter: &schema.PriceFilter{
					StartUsageAmount: strPtr("5"),
				},
			},
		},
	}
}

func computeSnapshotDiskSize(d *schema.ResourceData) *decimal.Decimal {
	if len(d.References("source_disk")) > 0 {
		return computeDiskSize(d.References("source_disk")[0])
	}

	return nil
}
