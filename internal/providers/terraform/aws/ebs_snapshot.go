package aws

import (
	"github.com/infracost/infracost/internal/config"
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

func NewEBSSnapshot(ctx *config.RunContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	gbVal := decimal.NewFromInt(int64(defaultVolumeSize))

	volumeRefs := d.References("volume_id")
	if len(volumeRefs) > 0 {
		if volumeRefs[0].Get("size").Exists() {
			gbVal = decimal.NewFromFloat(volumeRefs[0].Get("size").Float())
		}
	}

	var listBlockRequests *decimal.Decimal
	if u != nil && u.Get("monthly_list_block_requests").Exists() {
		listBlockRequests = decimalPtr(decimal.NewFromInt(u.Get("monthly_list_block_requests").Int()))
	}

	var getSnapshotBlockRequests *decimal.Decimal
	if u != nil && u.Get("monthly_get_block_requests").Exists() {
		getSnapshotBlockRequests = decimalPtr(decimal.NewFromInt(u.Get("monthly_get_block_requests").Int()))
	}

	var putSnapshotBlockRequests *decimal.Decimal
	if u != nil && u.Get("monthly_put_block_requests").Exists() {
		putSnapshotBlockRequests = decimalPtr(decimal.NewFromInt(u.Get("monthly_put_block_requests").Int()))
	}

	var fastSnapshotRestoreHours *decimal.Decimal
	if u != nil && u.Get("monthly_put_block_requests").Exists() {
		fastSnapshotRestoreHours = decimalPtr(decimal.NewFromInt(u.Get("fast_snapshot_restore_hours").Int()))
	}

	costComponents := []*schema.CostComponent{
		ebsSnapshotCostComponent(region, gbVal),
		{
			Name:            "Fast snapshot restore",
			Unit:            "DSU-hours",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: fastSnapshotRestoreHours,
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
			Name:            "ListChangedBlocks & ListSnapshotBlocks API requests",
			Unit:            "1k requests",
			UnitMultiplier:  decimal.NewFromInt(1000),
			MonthlyQuantity: listBlockRequests,
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
			Name:            "GetSnapshotBlock API requests",
			Unit:            "1k SnapshotAPIUnits",
			UnitMultiplier:  decimal.NewFromInt(1000),
			MonthlyQuantity: getSnapshotBlockRequests,
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
			Name:            "PutSnapshotBlock API requests",
			Unit:            "1k SnapshotAPIUnits",
			UnitMultiplier:  decimal.NewFromInt(1000),
			MonthlyQuantity: putSnapshotBlockRequests,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonEC2"),
				ProductFamily: strPtr("EBS direct API Requests"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/EBS:directAPI.snapshot.Put$/")},
				},
			},
		}}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func ebsSnapshotCostComponent(region string, gbVal decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "EBS snapshot storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
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
	}
}
