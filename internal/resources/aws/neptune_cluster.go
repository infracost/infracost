package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/shopspring/decimal"
)

type NeptuneCluster struct {
	Address               *string
	Region                *string
	BackupRetentionPeriod *int64
	StorageGb             *float64 `infracost_usage:"storage_gb"`
	MonthlyIoRequests     *int64   `infracost_usage:"monthly_io_requests"`
	BackupStorageGb       *float64 `infracost_usage:"backup_storage_gb"`
}

var NeptuneClusterUsageSchema = []*schema.UsageItem{{Key: "storage_gb", ValueType: schema.Float64, DefaultValue: 0}, {Key: "monthly_io_requests", ValueType: schema.Int64, DefaultValue: 0}, {Key: "backup_storage_gb", ValueType: schema.Float64, DefaultValue: 0}}

func (r *NeptuneCluster) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *NeptuneCluster) BuildResource() *schema.Resource {
	var storageGb, monthlyIoRequests *decimal.Decimal
	costComponents := make([]*schema.CostComponent, 0)

	if r.StorageGb != nil {
		storageGb = decimalPtr(decimal.NewFromFloat(*r.StorageGb))
	}
	if r.MonthlyIoRequests != nil {
		monthlyIoRequests = decimalPtr(decimal.NewFromInt(*r.MonthlyIoRequests))
	}

	costComponents = append(costComponents, r.storageIOsCostComponents("Storage", "GB", "StorageUsage", storageGb, 1))
	costComponents = append(costComponents, r.storageIOsCostComponents("I/O requests", "1M request", "StorageIOUsage", monthlyIoRequests, 1000000))
	var retentionPeriod *decimal.Decimal
	if r.BackupRetentionPeriod != nil {
		retentionPeriod = decimalPtr(decimal.NewFromInt(*r.BackupRetentionPeriod))
		if retentionPeriod.GreaterThan(decimal.NewFromInt(1)) {
			costComponents = append(costComponents, r.backupCostComponent())
		}
	}

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: NeptuneClusterUsageSchema,
	}
}

func (r *NeptuneCluster) storageIOsCostComponents(name, unit, usageType string, quantity *decimal.Decimal, unitMulti int) *schema.CostComponent {
	return &schema.CostComponent{

		Name:            name,
		Unit:            unit,
		UnitMultiplier:  decimal.NewFromInt(int64(unitMulti)),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     r.Region,
			Service:    strPtr("AmazonNeptune"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s$/i", usageType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func (r *NeptuneCluster) backupCostComponent() *schema.CostComponent {
	var backupStorage *decimal.Decimal
	if r.BackupStorageGb != nil {
		backupStorage = decimalPtr(decimal.NewFromFloat(*r.BackupStorageGb))
	}

	return &schema.CostComponent{
		Name:            "Backup storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: backupStorage,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     r.Region,
			Service:    strPtr("AmazonNeptune"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/BackupUsage$/i")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}
