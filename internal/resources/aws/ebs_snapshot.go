package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type EBSSnapshot struct {
	Address                  string
	Region                   string
	SizeGB                   *float64
	MonthlyListBlockRequests *int64 `infracost_usage:"monthly_list_block_requests"`
	MonthlyGetBlockRequests  *int64 `infracost_usage:"monthly_get_block_requests"`
	MonthlyPutBlockRequests  *int64 `infracost_usage:"monthly_put_block_requests"`
	FastSnapshotRestoreHours *int64 `infracost_usage:"fast_snapshot_restore_hours"`
}

func (r *EBSSnapshot) CoreType() string {
	return "EBSSnapshot"
}

func (r *EBSSnapshot) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{{Key: "monthly_list_block_requests", ValueType: schema.Int64, DefaultValue: 0}, {Key: "monthly_get_block_requests", ValueType: schema.Int64, DefaultValue: 0}, {Key: "monthly_put_block_requests", ValueType: schema.Int64, DefaultValue: 0}, {Key: "fast_snapshot_restore_hours", ValueType: schema.Int64, DefaultValue: 0}}
}

func (r *EBSSnapshot) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *EBSSnapshot) BuildResource() *schema.Resource {
	region := r.Region

	gbVal := decimal.NewFromFloat(float64(defaultVolumeSize))

	if r.SizeGB != nil {
		gbVal = decimal.NewFromFloat(*r.SizeGB)
	}

	var listBlockRequests *decimal.Decimal
	if r.MonthlyListBlockRequests != nil {
		listBlockRequests = decimalPtr(decimal.NewFromInt(*r.MonthlyListBlockRequests))
	}

	var getSnapshotBlockRequests *decimal.Decimal
	if r.MonthlyGetBlockRequests != nil {
		getSnapshotBlockRequests = decimalPtr(decimal.NewFromInt(*r.MonthlyGetBlockRequests))
	}

	var putSnapshotBlockRequests *decimal.Decimal
	if r.MonthlyPutBlockRequests != nil {
		putSnapshotBlockRequests = decimalPtr(decimal.NewFromInt(*r.MonthlyPutBlockRequests))
	}

	var fastSnapshotRestoreHours *decimal.Decimal
	if r.MonthlyPutBlockRequests != nil {
		fastSnapshotRestoreHours = decimalPtr(decimal.NewFromInt(*r.FastSnapshotRestoreHours))
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
			UsageBased: true,
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
			UsageBased: true,
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
			UsageBased: true,
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
			UsageBased: true,
		}}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents, UsageSchema: r.UsageSchema(),
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
		UsageBased: true,
	}
}
