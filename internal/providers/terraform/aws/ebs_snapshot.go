package aws

import (
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetEBSSnapshotRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_ebs_snapshot",
		RFunc:               NewEBSSnapshot,
		ReferenceAttributes: []string{"volume_id"},
	}
}

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
			Name:            "EBS snapshot storage",
			Unit:            "GB-months",
			UnitMultiplier:  1,
			MonthlyQuantity: &gbVal,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonEC2"),
				ProductFamily: strPtr("Storage Snapshot"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/EBS:SnapshotUsage$/")},
				},
			},
		},
		{
			Name:           "Fast snapshot restore",
			Unit:           "DSU-hours",
			UnitMultiplier: 1,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonEC2"),
				ProductFamily: strPtr("Fast Snapshot Restore"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/EBS:FastSnapshotRestore$/")},
				},
			},
		},
		{
			Name:           "ListChangedBlocks & ListSnapshotBlocks API requests",
			Unit:           "requests",
			UnitMultiplier: 1000,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonEC2"),
				ProductFamily: strPtr("EBS direct API Requests"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/EBS:directAPI.snapshot.List$/")},
				},
			},
		},
		{
			Name:           "GetSnapshotBlock API requests",
			Unit:           "SnapshotAPIUnits",
			UnitMultiplier: 1000,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonEC2"),
				ProductFamily: strPtr("EBS direct API Requests"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/EBS:directAPI.snapshot.Get$/")},
				},
			},
		},
		{
			Name:           "PutSnapshotBlock API requests",
			Unit:           "SnapshotAPIUnits",
			UnitMultiplier: 1000,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonEC2"),
				ProductFamily: strPtr("EBS direct API Requests"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/EBS:directAPI.snapshot.Put$/")},
				},
			},
		},
	}
}
