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
	StorageGb             *int64 `infracost_usage:"storage_gb"`
	MonthlyIoRequests     *int64 `infracost_usage:"monthly_io_requests"`
	BackupStorageGb       *int64 `infracost_usage:"backup_storage_gb"`
}

var NeptuneClusterUsageSchema = []*schema.UsageItem{{Key: "storage_gb", ValueType: schema.Int64, DefaultValue: 0}, {Key: "monthly_io_requests", ValueType: schema.Int64, DefaultValue: 0}, {Key: "backup_storage_gb", ValueType: schema.Int64, DefaultValue: 0}}

func (r *NeptuneCluster) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *NeptuneCluster) BuildResource() *schema.Resource {
	var storageGb, monthlyIoRequests *decimal.Decimal
	region := *r.Region
	costComponents := make([]*schema.CostComponent, 0)

	if r != nil && r.StorageGb != nil {
		storageGb = decimalPtr(decimal.NewFromInt(*r.StorageGb))
	}
	if r != nil && r.MonthlyIoRequests != nil {
		monthlyIoRequests = decimalPtr(decimal.NewFromInt(*r.MonthlyIoRequests))
	}

	costComponents = append(costComponents, neptuneClusterStorageIOsCostComponents("Storage", "GB", region, "StorageUsage", storageGb, 1))
	costComponents = append(costComponents, neptuneClusterStorageIOsCostComponents("I/O requests", "1M request", region, "StorageIOUsage", monthlyIoRequests, 1000000))
	var retentionPeriod *decimal.Decimal
	if r.BackupRetentionPeriod != nil {
		retentionPeriod = decimalPtr(decimal.NewFromInt(*r.BackupRetentionPeriod))
		if retentionPeriod.GreaterThan(decimal.NewFromInt(1)) {
			costComponents = append(costComponents, backupCostComponent(r, region))
		}
	}

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: NeptuneClusterUsageSchema,
	}
}
func backupCostComponent(r *NeptuneCluster, region string) *schema.CostComponent {

	var backupStorage *decimal.Decimal
	if r != nil && r.BackupStorageGb != nil {
		backupStorage = decimalPtr(decimal.NewFromInt(*r.BackupStorageGb))
	}

	return neptuneClusterBackupCostComponent(region, backupStorage)
}
func neptuneClusterStorageIOsCostComponents(name, unit, region, usageType string, quantity *decimal.Decimal, unitMulti int) *schema.CostComponent {
	return &schema.CostComponent{

		Name:            name,
		Unit:            unit,
		UnitMultiplier:  decimal.NewFromInt(int64(unitMulti)),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
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

func neptuneClusterBackupCostComponent(region string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{

		Name:            "Backup storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
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
