package aws

import (
	"infracost/pkg/schema"

	"github.com/shopspring/decimal"
)

func NewEBSSnapshot(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	gbVal := decimal.NewFromInt(int64(defaultVolumeSize))

	volumeRefs := d.References("volume_id")
	if len(volumeRefs) > 0 {
		if volumeRefs[0].Get("size").Exists() {
			gbVal = decimal.NewFromFloat(volumeRefs[0].Get("size").Float())
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: ebsSnapshotCostComponents(region, gbVal),
	}
}

func ebsSnapshotCostComponents(region string, gbVal decimal.Decimal) []*schema.CostComponent {
	return []*schema.CostComponent{
		{
			Name:            "Storage",
			Unit:            "GB-months",
			MonthlyQuantity: &gbVal,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonEC2"),
				ProductFamily: strPtr("Storage Snapshot"),
				AttributeFilters: &[]schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/EBS:SnapshotUsage$/")},
				},
			},
		},
	}
}
