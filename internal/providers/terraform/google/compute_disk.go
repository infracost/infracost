package google

import (
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetComputeDiskRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "google_compute_disk",
		RFunc:               NewComputeDisk,
		ReferenceAttributes: []string{"image", "snapshot"},
	}
}

func NewComputeDisk(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	zone := d.Get("zone").String()
	if zone != "" {
		region = zoneToRegion(zone)
	}

	diskType := d.Get("type").String()
	size := computeDiskSize(d)

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			computeDisk(region, diskType, size),
		},
	}
}

func computeDisk(region string, diskType string, size *decimal.Decimal) *schema.CostComponent {
	diskTypeDesc := "/^Storage PD Capacity/"
	diskTypeLabel := "Standard provisioned storage (pd-standard)"
	switch diskType {
	case "pd-balanced":
		diskTypeDesc = "/^Balanced PD Capacity/"
		diskTypeLabel = "Balanced provisioned storage (pd-balanced)"
	case "pd-ssd":
		diskTypeDesc = "/^SSD backed PD Capacity/"
		diskTypeLabel = "SSD provisioned storage (pd-ssd)"
	}

	return &schema.CostComponent{
		Name:            diskTypeLabel,
		Unit:            "GiB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: size,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(region),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: strPtr(diskTypeDesc)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			EndUsageAmount: strPtr(""), // use the non-free tier
		},
	}
}

func computeDiskSize(d *schema.ResourceData) *decimal.Decimal {
	if d.Get("size").Exists() {
		return decimalPtr(decimal.NewFromFloat(d.Get("size").Float()))
	}

	if len(d.References("image")) > 0 {
		return computeImageDiskSize(d.References("image")[0])
	}

	if len(d.References("snapshot")) > 0 {
		return computeSnapshotDiskSize(d.References("snapshot")[0])
	}

	return defaultDiskSize(d.Get("type").String())
}

func defaultDiskSize(diskType string) *decimal.Decimal {
	diskType = strings.ToLower(diskType)
	if diskType == "pd-balanced" || diskType == "pd-ssd" {
		return decimalPtr(decimal.NewFromInt(100))
	}
	return decimalPtr(decimal.NewFromInt(500))
}
